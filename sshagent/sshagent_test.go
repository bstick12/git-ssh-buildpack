package sshagent_test

import (
	"errors"
	"github.com/bstick12/git-ssh-buildpack/sshagent"
	"github.com/bstick12/git-ssh-buildpack/utils"
	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/libcfbuildpack/test"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"
)

//go:generate mockgen -source=sshagent.go -destination=sshagent_mocks.go -package=sshagent

func TestUnitDetect(t *testing.T) {

	spec.Run(t, "Contributed", testSshAgent, spec.Report(report.Terminal{}))

}

func testSshAgent(t *testing.T, when spec.G, it spec.S) {

	var (
		factory *test.BuildFactory
		mockRunner  *sshagent.MockRunner
		mockCtrl    *gomock.Controller
	)

	it.Before(func() {
		RegisterTestingT(t)
		factory = test.NewBuildFactory(t)

		mockCtrl = gomock.NewController(t)
		mockRunner = sshagent.NewMockRunner(mockCtrl)

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
			err := sshagent.Contribute(factory.Build, mockRunner)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("no GIT_SSH_KEY environment variable found"))

		})
	})

	when("GIT_SSH_KEY is available", func() {

		var sshkey = "VALUE"

		it("should execute required commands", func() {
			defer utils.ResetEnv(os.Environ())
			os.Clearenv()
			os.Setenv("GIT_SSH_KEY", sshkey)

			mockRunner.EXPECT().Run( ioutil.Discard, os.Stderr, nil, "ssh-agent","-a", sshagent.SshAgentSockAddress)
			mockRunner.EXPECT().Run( os.Stdout, os.Stderr, strings.NewReader(sshkey + "\n"), "ssh-add","-")
			mockRunner.EXPECT().Run( os.Stdout, os.Stderr, nil, "git","config", "--global", "url.git@github.com:.insteadOf", "https://github.com/")
			mockRunner.EXPECT().Run( os.Stdout, os.Stderr, nil, "ssh","-o", "StrictHostKeyChecking=accept-new", "git@github.com")
			err := sshagent.Contribute(factory.Build, mockRunner)
			Expect(err).To(BeNil())

		})

		when("commands fail", func() {
			it("should fail ssh-agent", func() {
				ret := errors.New("ssh-agent failed to start")
				mockRunner.EXPECT().Run(ioutil.Discard, os.Stderr, nil, "ssh-agent", "-a", sshagent.SshAgentSockAddress).Return(ret)

				defer utils.ResetEnv(os.Environ())
				os.Clearenv()
				os.Setenv("GIT_SSH_KEY", sshkey)

				err := sshagent.Contribute(factory.Build, mockRunner)
				Expect(err).To(Equal(ret))
			})

			it("should fail ssh-add", func() {
				ret := errors.New("ssh-add failed to start")
				mockRunner.EXPECT().Run( ioutil.Discard, os.Stderr, nil, "ssh-agent","-a", sshagent.SshAgentSockAddress)
				mockRunner.EXPECT().Run( os.Stdout, os.Stderr, strings.NewReader(sshkey + "\n"), "ssh-add","-").Return(ret)

				defer utils.ResetEnv(os.Environ())
				os.Clearenv()
				os.Setenv("GIT_SSH_KEY", sshkey)

				err := sshagent.Contribute(factory.Build, mockRunner)
				Expect(err).To(Equal(ret))
			})

			it("should fail git", func() {
				ret := errors.New("git failed to start")
				mockRunner.EXPECT().Run( ioutil.Discard, os.Stderr, nil, "ssh-agent","-a", sshagent.SshAgentSockAddress)
				mockRunner.EXPECT().Run( os.Stdout, os.Stderr, strings.NewReader(sshkey + "\n"), "ssh-add","-")
				mockRunner.EXPECT().Run( os.Stdout, os.Stderr, nil, "git","config", "--global", "url.git@github.com:.insteadOf", "https://github.com/").Return(ret)

				defer utils.ResetEnv(os.Environ())
				os.Clearenv()
				os.Setenv("GIT_SSH_KEY", sshkey)

				err := sshagent.Contribute(factory.Build, mockRunner)
				Expect(err).To(Equal(ret))
			})

			it("should fail ssh", func() {
				// Need a failure with an exit code over 1
				ret := exec.Command("ssh", "unknown.example.com").Run()

				mockRunner.EXPECT().Run( ioutil.Discard, os.Stderr, nil, "ssh-agent","-a", sshagent.SshAgentSockAddress)
				mockRunner.EXPECT().Run( os.Stdout, os.Stderr, strings.NewReader(sshkey + "\n"), "ssh-add","-")
				mockRunner.EXPECT().Run( os.Stdout, os.Stderr, nil, "git","config", "--global", "url.git@github.com:.insteadOf", "https://github.com/")
				mockRunner.EXPECT().Run( os.Stdout, os.Stderr, nil, "ssh","-o", "StrictHostKeyChecking=accept-new", "git@github.com").Return(ret)

				defer utils.ResetEnv(os.Environ())
				os.Clearenv()
				os.Setenv("GIT_SSH_KEY", sshkey)

				err := sshagent.Contribute(factory.Build, mockRunner)
				Expect(err).To(Equal(ret))
			})

		})
	})

}


