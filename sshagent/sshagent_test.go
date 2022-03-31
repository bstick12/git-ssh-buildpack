package sshagent_test

import (
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/avarteqgmbh/git-ssh-buildpack/sshagent"
	"github.com/avarteqgmbh/git-ssh-buildpack/utils"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/scribe"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

//go:generate mockgen -source=sshagent.go -destination=sshagent_mocks.go -package=sshagent

func TestUnitDetect(t *testing.T) {
	spec.Run(t, "SSH Agent", testSshAgent, spec.Report(report.Terminal{}))
}

func testSshAgent(t *testing.T, when spec.G, it spec.S) {
	Expect := NewWithT(t).Expect

	var (
		mockRunner *sshagent.MockRunner
		mockCtrl   *gomock.Controller
		logger     scribe.Emitter
		context    packit.BuildContext
	)

	it.Before(func() {
		mockCtrl = gomock.NewController(t)
		mockRunner = sshagent.NewMockRunner(mockCtrl)
		logger = scribe.NewEmitter(os.Stdout).WithLevel(os.Getenv("BP_LOG_LEVEL"))
		context = packit.BuildContext{}

	})

	when("GIT_SSH_KEY is not available", func() {
		it("should return err", func() {
			defer utils.ResetEnv(os.Environ())
			os.Clearenv()
			_, err := sshagent.Contribute(context, logger, mockRunner)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("No GIT_SSH_KEY environment variable found"))
		})
	})

	when("GIT_SSH_KEY is available", func() {
		var (
			sshkey   = "VALUE"
			sshHosts = "bitbucket.org,gitlab.com"
		)

		it("should execute required commands", func() {
			defer utils.ResetEnv(os.Environ())
			os.Clearenv()
			os.Setenv("GIT_SSH_KEY", sshkey)

			mockRunner.EXPECT().Run(ioutil.Discard, os.Stderr, nil, "ssh-agent", "-a", sshagent.SockAddress)
			mockRunner.EXPECT().Run(os.Stdout, os.Stderr, strings.NewReader(sshkey+"\n"), "ssh-add", "-")
			mockRunner.EXPECT().Run(os.Stdout, os.Stderr, nil, "git", "config", "--global", "url.git@github.com:.insteadOf", "https://github.com/")

			mockRunner.EXPECT().Run(os.Stdout, os.Stderr, nil, "ssh", "-o", "StrictHostKeyChecking=accept-new", "git@github.com")
			_, err := sshagent.Contribute(context, logger, mockRunner)
			Expect(err).To(BeNil())
		})

		when("GIT_SSH_HOSTS is set", func() {
			it("should execute the required commands", func() {
				defer utils.ResetEnv(os.Environ())
				os.Clearenv()
				os.Setenv("GIT_SSH_KEY", sshkey)
				os.Setenv("GIT_SSH_HOSTS", sshHosts)

				mockRunner.EXPECT().Run(ioutil.Discard, os.Stderr, nil, "ssh-agent", "-a", sshagent.SockAddress)
				mockRunner.EXPECT().Run(os.Stdout, os.Stderr, strings.NewReader(sshkey+"\n"), "ssh-add", "-")
				mockRunner.EXPECT().Run(os.Stdout, os.Stderr, nil, "git", "config", "--global", "url.git@bitbucket.org:.insteadOf", "https://bitbucket.org/")
				mockRunner.EXPECT().Run(os.Stdout, os.Stderr, nil, "git", "config", "--global", "url.git@gitlab.com:.insteadOf", "https://gitlab.com/")

				mockRunner.EXPECT().Run(os.Stdout, os.Stderr, nil, "ssh", "-o", "StrictHostKeyChecking=accept-new", "git@bitbucket.org")
				mockRunner.EXPECT().Run(os.Stdout, os.Stderr, nil, "ssh", "-o", "StrictHostKeyChecking=accept-new", "git@gitlab.com")

				_, err := sshagent.Contribute(context, logger, mockRunner)
				Expect(err).To(BeNil())
			})
		})

		when("commands fail", func() {
			it("should fail ssh-agent", func() {
				ret := errors.New("ssh-agent failed to start")
				mockRunner.EXPECT().Run(ioutil.Discard, os.Stderr, nil, "ssh-agent", "-a", sshagent.SockAddress).Return(ret)

				defer utils.ResetEnv(os.Environ())
				os.Clearenv()
				os.Setenv("GIT_SSH_KEY", sshkey)

				_, err := sshagent.Contribute(context, logger, mockRunner)
				Expect(err).To(Equal(ret))
			})

			it("should fail ssh-add", func() {
				ret := errors.New("ssh-add failed to start")
				mockRunner.EXPECT().Run(ioutil.Discard, os.Stderr, nil, "ssh-agent", "-a", sshagent.SockAddress)
				mockRunner.EXPECT().Run(os.Stdout, os.Stderr, strings.NewReader(sshkey+"\n"), "ssh-add", "-").Return(ret)

				defer utils.ResetEnv(os.Environ())
				os.Clearenv()
				os.Setenv("GIT_SSH_KEY", sshkey)

				_, err := sshagent.Contribute(context, logger, mockRunner)
				Expect(err).To(Equal(ret))
			})

			it("should fail git", func() {
				ret := errors.New("git failed to start")
				mockRunner.EXPECT().Run(ioutil.Discard, os.Stderr, nil, "ssh-agent", "-a", sshagent.SockAddress)
				mockRunner.EXPECT().Run(os.Stdout, os.Stderr, strings.NewReader(sshkey+"\n"), "ssh-add", "-")
				mockRunner.EXPECT().Run(os.Stdout, os.Stderr, nil, "git", "config", "--global", "url.git@github.com:.insteadOf", "https://github.com/").Return(ret)

				defer utils.ResetEnv(os.Environ())
				os.Clearenv()
				os.Setenv("GIT_SSH_KEY", sshkey)

				_, err := sshagent.Contribute(context, logger, mockRunner)
				Expect(err).To(Equal(ret))
			})

			it("should fail ssh", func() {
				// Need a failure with an exit code over 1
				ret := exec.Command("ssh", "unknown.example.com").Run()

				mockRunner.EXPECT().Run(ioutil.Discard, os.Stderr, nil, "ssh-agent", "-a", sshagent.SockAddress)
				mockRunner.EXPECT().Run(os.Stdout, os.Stderr, strings.NewReader(sshkey+"\n"), "ssh-add", "-")
				mockRunner.EXPECT().Run(os.Stdout, os.Stderr, nil, "git", "config", "--global", "url.git@github.com:.insteadOf", "https://github.com/")
				mockRunner.EXPECT().Run(os.Stdout, os.Stderr, nil, "ssh", "-o", "StrictHostKeyChecking=accept-new", "git@github.com").Return(ret)

				defer utils.ResetEnv(os.Environ())
				os.Clearenv()
				os.Setenv("GIT_SSH_KEY", sshkey)

				_, err := sshagent.Contribute(context, logger, mockRunner)
				Expect(err).To(Equal(ret))
			})
		})
	})
}
