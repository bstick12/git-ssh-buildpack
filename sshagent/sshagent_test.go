package sshagent_test

import (
	"fmt"
	"github.com/bstick12/git-key-buildpack/sshagent"
	"github.com/bstick12/git-key-buildpack/utils"
	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/libcfbuildpack/test"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestUnitDetect(t *testing.T) {

	spec.Run(t, "Contributed", testContribute, spec.Report(report.Terminal{}))

}

type CmdFunctionParams struct {
	Stdout io.Writer
	StdErr io.Writer
	Stdin io.Reader
	Command string
	Args []string
	Return error
}

type TestRunner struct {
	Runner func () error
}

func(tr *TestRunner) Run() error {
	return tr.Runner()
}

var cmdFunctions map[string]CmdFunctionParams

func testContribute(t *testing.T, when spec.G, it spec.S) {

	var factory *test.BuildFactory

	it.Before(func() {
		RegisterTestingT(t)
		factory = test.NewBuildFactory(t)

		factory.Build.BuildPlan = buildplan.BuildPlan{
			sshagent.Dependency: buildplan.Dependency{
				Metadata: buildplan.Metadata{
					"build": true,
				},
			},
		}
	})

	when("GIT_SSH_KEY is not available", func() {
		it("should return err", func() {

			defer utils.ResetEnv(os.Environ())
			os.Clearenv()
			err := sshagent.Contribute(factory.Build)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("no GIT_SSH_KEY environment variable found"))

		})
	})

	when("GIT_SSH_KEY is available", func() {

		var sshkey = "value"

		it.Before(func() {
			cmdFunctions = make(map[string]CmdFunctionParams)
			cmdFunctions["ssh-agent"] = CmdFunctionParams{
				Stdout: ioutil.Discard,
				StdErr: os.Stderr,
				Stdin: nil,
				Args: []string{"-a", sshagent.SshAgentSockAddress},
				Return: nil,
			}
			cmdFunctions["ssh-add"] = CmdFunctionParams{
				Stdout: os.Stdout,
				StdErr: os.Stderr,
				Stdin: strings.NewReader(sshkey + "\n"),
				Args: []string{"-"},
				Return: nil,
			}
			cmdFunctions["git"] = CmdFunctionParams{
				Stdout: os.Stdout,
				StdErr: os.Stderr,
				Stdin: nil,
				Args: []string{"config", "--global", "url.git@github.com:.insteadOf", "https://github.com/"},
				Return: nil,
			}
			cmdFunctions["ssh"] = CmdFunctionParams{
				Stdout: os.Stdout,
				StdErr: os.Stderr,
				Stdin: nil,
				Args: []string{"-o", "StrictHostKeyChecking=accept-new", "git@github.com"},
				Return: nil,
			}})

		it("should execute required commands", func() {
			defer utils.ResetEnv(os.Environ())
			os.Clearenv()
			os.Setenv("GIT_SSH_KEY", sshkey)
			sshagent.CmdFunction = CmdSuccess
			err := sshagent.Contribute(factory.Build)
			Expect(err).To(BeNil())

		})

		when("commands fail", func() {
			it("should fail ssh-agent", func() {
				ret := errors.New("ssh-agent failed to start")
				changeCmdReturn("ssh-agent", ret)
				defer utils.ResetEnv(os.Environ())
				os.Clearenv()
				os.Setenv("GIT_SSH_KEY", sshkey)
				sshagent.CmdFunction = CmdFailure
				err := sshagent.Contribute(factory.Build)
				Expect(err).To(Equal(ret))
			})

			it("should fail ssh-add", func() {
				ret := errors.New("ssh-add failed to start")
				changeCmdReturn("ssh-add", ret)
				defer utils.ResetEnv(os.Environ())
				os.Clearenv()
				os.Setenv("GIT_SSH_KEY", sshkey)
				sshagent.CmdFunction = CmdFailure
				err := sshagent.Contribute(factory.Build)
				Expect(err).To(Equal(ret))
			})

			it("should fail git", func() {
				ret := errors.New("git failed to start")
				changeCmdReturn("git", ret)
				defer utils.ResetEnv(os.Environ())
				os.Clearenv()
				os.Setenv("GIT_SSH_KEY", sshkey)
				sshagent.CmdFunction = CmdFailure
				err := sshagent.Contribute(factory.Build)
				Expect(err).To(Equal(ret))
			})

			it("should fail ssh", func() {
				// Need a failure with an exit code over 1
				ret := exec.Command("ssh", "unknown.example.com").Run()

				changeCmdReturn("ssh", ret)
				defer utils.ResetEnv(os.Environ())
				os.Clearenv()
				os.Setenv("GIT_SSH_KEY", sshkey)
				sshagent.CmdFunction = CmdFailure
				err := sshagent.Contribute(factory.Build)
				Expect(err).To(Equal(ret))
			})

		})
	})

}

func changeCmdReturn(command string, ret error) {
	cmdFunction := cmdFunctions[command]
	cmdFunction.Return = ret
	cmdFunctions[command] = cmdFunction

}

func CmdSuccess (stdout, stderr io.Writer, stdin io.Reader, command string, args ...string) sshagent.Runner {
	return &TestRunner {
		Runner: func() error {
			cmdFunction, ok := cmdFunctions[command]
			Expect(ok).To(BeTrue(), fmt.Sprintf("Failed to find command %s", command))
			Expect(stdout).To(Equal(cmdFunction.Stdout))
			Expect(stderr).To(Equal(cmdFunction.StdErr))
			if cmdFunction.Stdin == nil {
				Expect(stdin).To(BeNil())
			} else {
				Expect(stdin).To(Equal(cmdFunction.Stdin))
			}
			Expect(args).To(Equal(cmdFunction.Args))
			return cmdFunction.Return
		},
	}
}

func CmdFailure (_, _ io.Writer, _ io.Reader, command string, _ ...string) sshagent.Runner {
	return &TestRunner{
		Runner: func() error {
			cmdFunction, ok := cmdFunctions[command]
			Expect(ok).To(BeTrue(), fmt.Sprintf("Failed to find command %s", command))
			return cmdFunction.Return
		},
	}
}



