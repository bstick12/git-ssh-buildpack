package main_test

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	cmdBuild "github.com/avarteqgmbh/git-ssh-buildpack/cmd/build"
	"github.com/avarteqgmbh/git-ssh-buildpack/sshagent"
	"github.com/avarteqgmbh/git-ssh-buildpack/utils"

	"github.com/buildpack/libbuildpack/buildpackplan"
	"github.com/cloudfoundry/libcfbuildpack/build"
	"github.com/cloudfoundry/libcfbuildpack/test"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitDetect(t *testing.T) {
	spec.Run(t, "Build", testDetect, spec.Report(report.Terminal{}))
}

func testDetect(t *testing.T, when spec.G, it spec.S) {
	var (
		factory    *test.BuildFactory
		mockRunner *sshagent.MockRunner
		mockCtrl   *gomock.Controller
	)

	it.Before(func() {
		RegisterTestingT(t)
		factory = test.NewBuildFactory(t)
		mockCtrl = gomock.NewController(t)
		mockRunner = sshagent.NewMockRunner(mockCtrl)
	})

	when("building source", func() {
		var sshKey = "VALUE"

		it("should pass if successful", func() {
			defer utils.ResetEnv(os.Environ())
			os.Clearenv()
			os.Setenv("GIT_SSH_KEY", sshKey)

			factory.AddPlan(buildpackplan.Plan{
				Name:     sshagent.Dependency,
				Version:  "",
				Metadata: buildpackplan.Metadata{"build": true},
			})

			mockRunner.EXPECT().Run(ioutil.Discard, os.Stderr, nil, "ssh-agent", "-a", sshagent.SockAddress)
			mockRunner.EXPECT().Run(os.Stdout, os.Stderr, strings.NewReader(sshKey+"\n"), "ssh-add", "-")
			mockRunner.EXPECT().Run(os.Stdout, os.Stderr, nil, "git", "config", "--global", "url.git@github.com:.insteadOf", "https://github.com/")
			mockRunner.EXPECT().Run(os.Stdout, os.Stderr, nil, "ssh", "-o", "StrictHostKeyChecking=accept-new", "git@github.com")

			code, err := cmdBuild.RunBuild(factory.Build, mockRunner)
			Expect(err).NotTo(HaveOccurred())
			Expect(code).To(Equal(build.SuccessStatusCode))

			sshAgentLayer := factory.Build.Layers.Layer(sshagent.Dependency)
			Expect(sshAgentLayer).To(test.HaveLayerMetadata(true, true, false))
		})

		it("should fail if it doesn't contribute", func() {
			defer utils.ResetEnv(os.Environ())
			os.Clearenv()
			os.Setenv("GIT_SSH_KEY", sshKey)

			code, err := cmdBuild.RunBuild(factory.Build, mockRunner)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Failed to find build plan"))
			Expect(code).To(Equal(cmdBuild.FailureStatusCode))
		})
	})
}
