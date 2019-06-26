package sshagent

import (
	"errors"
	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/libcfbuildpack/build"
	"github.com/cloudfoundry/libcfbuildpack/layers"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

const (
	Dependency =  "sshagent"
	SshAgentSockAddress = "/tmp/git-ssh-buildpack.sock"
)

type Runner interface {
 	Run(stdout, stderr io.Writer, stdin io.Reader, command string, args ...string) error
}



func Contribute(context build.Build, runner Runner) error {

	dependency, wantLayer := context.BuildPlan[Dependency]
	if !wantLayer {
		return errors.New("layer %s is not wanted")
	}

	layer := context.Layers.HelperLayer(Dependency, "SSH Agent Layer")
	sshkey, ok := os.LookupEnv("GIT_SSH_KEY")
	if !ok {
		layer.Logger.Error("No GIT_SSH_KEY environment variable found")
		return errors.New("no GIT_SSH_KEY environment variable found")
	}

	layer.Logger.SubsequentLine("Starting SSH agent")
	err := runner.Run(ioutil.Discard, os.Stderr, nil, "ssh-agent", "-a", SshAgentSockAddress)
	if err != nil {
		layer.Logger.Error("Failed to start ssh-agent [%v]", err)
		return err
	}

	os.Setenv("SSH_AUTH_SOCK", SshAgentSockAddress)
	err = runner.Run(os.Stdout, os.Stderr, strings.NewReader(sshkey +"\n"), "ssh-add", "-")
	if err != nil {
		layer.Logger.Error("Failed to add SSH Key [%v]", err)
		return err
	}

	err = runner.Run(os.Stdout, os.Stderr,nil, "git", "config", "--global", "url.git@github.com:.insteadOf","https://github.com/")
	if err != nil {
		layer.Logger.Error("Failed to configure git for SSH [%v]", err)
		return err
	}

	err = runner.Run(os.Stdout, os.Stderr,nil, "ssh", "-o", "StrictHostKeyChecking=accept-new", "git@github.com")
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				if status.ExitStatus()> 1 {
					layer.Logger.Error("Failed to authorize with github")
					return err
				}
			}
		}
	}

	var sshAgentHelperLayerContributor = func(artifact string, layer layers.HelperLayer) error {
		layer.AppendBuildEnv("SSH_AUTH_SOCK", "%s", SshAgentSockAddress)
		return nil
	}

	if err := layer.Contribute(sshAgentHelperLayerContributor, flags(dependency)...); err != nil {
		layer.Logger.Error("Failed to contribute helper layer [%v]", err)
		return err
	}

	return nil
}

func flags(plan buildplan.Dependency) []layers.Flag {
	var flags []layers.Flag
	cache, _ := plan.Metadata["cache"].(bool)
	if cache {
		flags = append(flags, layers.Cache)
	}
	build, _ := plan.Metadata["build"].(bool)
	if build {
		flags = append(flags, layers.Build)
	}
	launch, _ := plan.Metadata["launch"].(bool)
	if launch {
		flags = append(flags, layers.Launch)
	}
	return flags
}

type CmdRunner struct {}

func (nr CmdRunner) Run(stdout, stderr io.Writer, stdin io.Reader, command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.Stdin = stdin
	return cmd.Run()
}
