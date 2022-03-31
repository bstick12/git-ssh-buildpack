package main_test

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/scribe"

	cmd_run "github.com/avarteqgmbh/git-ssh-buildpack/run"
	"github.com/avarteqgmbh/git-ssh-buildpack/sshagent"
	"github.com/avarteqgmbh/git-ssh-buildpack/utils"
	"github.com/golang/mock/gomock"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitDetect(t *testing.T) {
	spec.Run(t, "Detect", testDetect, spec.Report(report.Terminal{}))
}

func testDetect(t *testing.T, when spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect
	)

	when("the os environment variable is present", func() {
		it("should add git-ssh-buildpack to the buildplan", func() {
			defer utils.ResetEnv(os.Environ())
			os.Clearenv()
			os.Setenv("GIT_SSH_KEY", "VALUE")
			detectFunc := cmd_run.Detect()
			detectResult, err := detectFunc(packit.DetectContext{})
			Expect(err).NotTo(HaveOccurred())
			Expect(detectResult.Plan).To(Equal(
				packit.BuildPlan{
					Provides: []packit.BuildPlanProvision{
						{Name: sshagent.Dependency},
					},
					Requires: []packit.BuildPlanRequirement{
						{
							Name: sshagent.Dependency,
							Metadata: cmd_run.GitSSHBuildPlanMetadata{
								Build:  true,
								Cache:  false,
								Launch: false,
							},
						},
					},
				},
			))
		})
	})

	when("the os environment variable is not present", func() {
		it("should not add git-ssh-buildpack to the buildplan", func() {
			defer utils.ResetEnv(os.Environ())
			os.Clearenv()
			detectFunc := cmd_run.Detect()
			detectResult, err := detectFunc(packit.DetectContext{})
			Expect(err).To(HaveOccurred())
			Expect(detectResult.Plan).To(Equal(packit.BuildPlan{Provides: nil, Requires: nil, Or: nil}))
		})
	})
}

func TestUnitBuild(t *testing.T) {
	spec.Run(t, "Build", testBuild, spec.Report(report.Terminal{}))
}

func testBuild(t *testing.T, when spec.G, it spec.S) {
	var (
		Expect     = NewWithT(t).Expect
		mockRunner *sshagent.MockRunner
		mockCtrl   *gomock.Controller
		logger     scribe.Emitter
	)

	it.Before(func() {
		RegisterTestingT(t)
		mockCtrl = gomock.NewController(t)
		mockRunner = sshagent.NewMockRunner(mockCtrl)
		logger = scribe.NewEmitter(os.Stdout)
	})

	when("building source", func() {
		var sshKey = "VALUE"

		it("should pass if successful", func() {
			defer utils.ResetEnv(os.Environ())
			os.Clearenv()
			os.Setenv("GIT_SSH_KEY", sshKey)

			mockRunner.EXPECT().Run(ioutil.Discard, os.Stderr, nil, "ssh-agent", "-a", sshagent.SockAddress)
			mockRunner.EXPECT().Run(os.Stdout, os.Stderr, strings.NewReader(sshKey+"\n"), "ssh-add", "-")
			mockRunner.EXPECT().Run(os.Stdout, os.Stderr, nil, "git", "config", "--global", "url.git@github.com:.insteadOf", "https://github.com/")
			mockRunner.EXPECT().Run(os.Stdout, os.Stderr, nil, "ssh", "-o", "StrictHostKeyChecking=accept-new", "git@github.com")

			context := packit.BuildContext{
				WorkingDir: "some-dir",
				CNBPath:    "some-cnb-dir",
				Stack:      "some-stack",
				BuildpackInfo: packit.BuildpackInfo{
					Name:    "Some Buildpack",
					Version: "some-version",
				},
				Layers: packit.Layers{Path: "."},
			}
			buildResult, err := sshagent.Contribute(context, logger, mockRunner)
			Expect(err).NotTo(HaveOccurred())
			Expect(buildResult).To(Equal(
				packit.BuildResult{
					Layers: []packit.Layer{
						{
							Path:      "sshagent",
							Name:      "sshagent",
							Build:     false,
							Launch:    true,
							Cache:     false,
							SharedEnv: packit.Environment{},
							BuildEnv: packit.Environment{
								"SSH_AUTH_SOCK.default": "/tmp/git-ssh-buildpack.sock",
							},
							LaunchEnv:        packit.Environment{},
							ProcessLaunchEnv: map[string]packit.Environment{},
							Metadata:         nil,
							SBOM:             nil,
						},
					},
				},
			))
		})
	})
}
