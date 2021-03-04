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
	// Dependency carries the name of the dependency this Buildpack offers
	Dependency = "sshagent"

	// SockAddress is the default address of the socket of the SSH agent started during build
	SockAddress = "/tmp/git-ssh-buildpack.sock"
)

// Runner is an interface used during build and unit testing
type Runner interface {
	Run(stdout, stderr io.Writer, stdin io.Reader, command string, args ...string) error
}

// Contribute adds the logic this Buildpack contributes to a build
func Contribute(context build.Build, runner Runner) error {
	dependency, wantDependency, err := context.Plans.GetShallowMerged(Dependency)
	if err != nil || !wantDependency {
		return fmt.Errorf("layer %s is not wanted", Dependency)
	}

	layer := context.Layers.HelperLayer(Dependency, "SSH Agent Layer")
	sshkey, ok := os.LookupEnv("GIT_SSH_KEY")
	if !ok {
		layer.Logger.HeaderError("No GIT_SSH_KEY environment variable found")
		return errors.New("No GIT_SSH_KEY environment variable found")
	}

	layer.Logger.Body("Starting SSH agent")
	err = runner.Run(ioutil.Discard, os.Stderr, nil, "ssh-agent", "-a", SockAddress)
	if err != nil {
		layer.Logger.HeaderError("Failed to start ssh-agent [%v]", err)
		return err
	}

	os.Setenv("SSH_AUTH_SOCK", SockAddress)
	err = runner.Run(os.Stdout, os.Stderr, strings.NewReader(sshkey+"\n"), "ssh-add", "-")
	if err != nil {
		layer.Logger.HeaderError("Failed to add SSH Key [%v]", err)
		return err
	}

	for _, host := range getGitSSHHosts() {
		layer.Logger.Body("Configuring host [%s]", host)
		err = runner.Run(os.Stdout, os.Stderr, nil, "git", "config", "--global",
			fmt.Sprintf("url.git@%s:.insteadOf", host), fmt.Sprintf("https://%s/", host))
		if err != nil {
			layer.Logger.HeaderError("Failed to configure git for SSH on host [%s] [%v]", host, err)
			return err
		}

		err = runner.Run(os.Stdout, os.Stderr, nil, "ssh", "-o", "StrictHostKeyChecking=accept-new",
			fmt.Sprintf("git@%s", host))
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
					if status.ExitStatus() > 1 {
						layer.Logger.BodyError("Failed to authorize with [%s]", host)
						return err
					}
				}
			}
		}
	}

	var sshAgentHelperLayerContributor = func(artifact string, layer layers.HelperLayer) error {
		err = layer.AppendBuildEnv("SSH_AUTH_SOCK", "%s", SockAddress)
		if err != nil {
			return err
		}
		return nil
	}

	if err := layer.Contribute(sshAgentHelperLayerContributor, flags(dependency)...); err != nil {
		layer.Logger.BodyError("Failed to contribute helper layer [%v]", err)
		return err
	}

	return nil
}

func getGitSSHHosts() []string {
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

// CmdRunner is used to run commands
type CmdRunner struct{}

// Run a particular command and intercept stdin, stdout and stderr
func (nr CmdRunner) Run(stdout, stderr io.Writer, stdin io.Reader, command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.Stdin = stdin
	return cmd.Run()
}
