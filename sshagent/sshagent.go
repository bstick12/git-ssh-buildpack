package sshagent

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/buildpack/libbuildpack/buildpackplan"
	"github.com/cloudfoundry/libcfbuildpack/build"
	"github.com/cloudfoundry/libcfbuildpack/layers"
)

const (
	Dependency  = "sshagent"
	SockAddress = "/tmp/git-ssh-buildpack.sock"
)

type Runner interface {
	Run(stdout, stderr io.Writer, stdin io.Reader, command string, args ...string) error
}

func Contribute(context build.Build, runner Runner) error {
	dependency, wantDependency, err := context.Plans.GetShallowMerged(Dependency)
	if err != nil || !wantDependency {
		return errors.New(fmt.Sprintf("layer %s is not wanted", Dependency))
	}

	layer := context.Layers.HelperLayer(Dependency, "SSH Agent Layer")
	sshkey, ok := os.LookupEnv("GIT_SSH_KEY")
	if !ok {
		layer.Logger.Error("No GIT_SSH_KEY environment variable found")
		return errors.New("No GIT_SSH_KEY environment variable found")
	}

	layer.Logger.SubsequentLine("Starting SSH agent")
	err = runner.Run(ioutil.Discard, os.Stderr, nil, "ssh-agent", "-a", SockAddress)
	if err != nil {
		layer.Logger.Error("Failed to start ssh-agent [%v]", err)
		return err
	}

	os.Setenv("SSH_AUTH_SOCK", SockAddress)
	err = runner.Run(os.Stdout, os.Stderr, strings.NewReader(sshkey+"\n"), "ssh-add", "-")
	if err != nil {
		layer.Logger.Error("Failed to add SSH Key [%v]", err)
		return err
	}

	for _, host := range getGitSshHosts() {
		layer.Logger.SubsequentLine("Configuring host [%s]", host)
		err = runner.Run(os.Stdout, os.Stderr, nil, "git", "config", "--global",
			fmt.Sprintf("url.git@%s:.insteadOf", host), fmt.Sprintf("https://%s/", host))
		if err != nil {
			layer.Logger.Error("Failed to configure git for SSH on host [%s] [%v]", host, err)
			return err
		}

		err = runner.Run(os.Stdout, os.Stderr, nil, "ssh", "-o", "StrictHostKeyChecking=accept-new",
			fmt.Sprintf("git@%s", host))
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
					if status.ExitStatus() > 1 {
						layer.Logger.Error("Failed to authorize with [%s]", host)
						return err
					}
				}
			}
		}
	}

	var sshAgentHelperLayerContributor = func(artifact string, layer layers.HelperLayer) error {
		layer.AppendBuildEnv("SSH_AUTH_SOCK", "%s", SockAddress)
		return nil
	}

	if err := layer.Contribute(sshAgentHelperLayerContributor, flags(dependency)...); err != nil {
		layer.Logger.Error("Failed to contribute helper layer [%v]", err)
		return err
	}

	return nil
}

func getGitSshHosts() []string {
	if value, ok := os.LookupEnv("GIT_SSH_HOSTS"); ok {
		return strings.Split(value, ",")
	}
	return []string{"github.com"}
}

func flags(plan buildpackplan.Plan) []layers.Flag {
	flags := []layers.Flag{layers.Cache}

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

type CmdRunner struct{}

func (nr CmdRunner) Run(stdout, stderr io.Writer, stdin io.Reader, command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.Stdin = stdin
	return cmd.Run()
}
