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

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/scribe"
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
func Contribute(context packit.BuildContext, logger scribe.Emitter, runner Runner) (packit.BuildResult, error) {

	logger.Title("%s %s", context.BuildpackInfo.Name, context.BuildpackInfo.Version)

	assetsLayer, err := context.Layers.Get(Dependency)
	if err != nil {
		return packit.BuildResult{}, err
	}
	logger.Debug.Subprocess(assetsLayer.Path)
	logger.Debug.Break()

	sshkey, ok := os.LookupEnv("GIT_SSH_KEY")
	if !ok {
		logger.Process("No GIT_SSH_KEY environment variable found")
		return packit.BuildResult{}, errors.New("No GIT_SSH_KEY environment variable found")
	}

	logger.Subprocess("Starting SSH agent")
	err = runner.Run(ioutil.Discard, os.Stderr, nil, "ssh-agent", "-a", SockAddress)
	if err != nil {
		logger.Subprocess("Failed to start ssh-agent [%v]", err)
		return packit.BuildResult{}, err
	}

	os.Setenv("SSH_AUTH_SOCK", SockAddress)
	err = runner.Run(os.Stdout, os.Stderr, strings.NewReader(sshkey+"\n"), "ssh-add", "-")
	if err != nil {
		logger.Subprocess("Failed to add SSH Key [%v]", err)
		return packit.BuildResult{}, err
	}

	for _, host := range getGitSSHHosts() {
		logger.Subprocess("Configuring host [%s]", host)
		err = runner.Run(os.Stdout, os.Stderr, nil, "git", "config", "--global",
			fmt.Sprintf("url.git@%s:.insteadOf", host), fmt.Sprintf("https://%s/", host))
		if err != nil {
			logger.Subprocess("Failed to configure git for SSH on host [%s] [%v]", host, err)
			return packit.BuildResult{}, err
		}

		_, ok = os.LookupEnv("GIT_SSH_DONT_CONNECT")
		if !ok {
			err = runner.Run(os.Stdout, os.Stderr, nil, "ssh", "-o", "StrictHostKeyChecking=accept-new",
				fmt.Sprintf("git@%s", host))
			if err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
						if status.ExitStatus() > 1 {
							logger.Subprocess("Failed to authorize with [%s]", host)
							return packit.BuildResult{}, err
						}
					}
				}
			}
		}
	}

	assetsLayer.BuildEnv.Default("SSH_AUTH_SOCK", SockAddress)
	assetsLayer.Launch = true

	return packit.BuildResult{
		Layers: []packit.Layer{assetsLayer},
	}, nil
}

func getGitSSHHosts() []string {
	if value, ok := os.LookupEnv("GIT_SSH_HOSTS"); ok {
		return strings.Split(value, ",")
	}
	return []string{"github.com"}
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
