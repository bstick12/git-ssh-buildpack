package main_test

import (
	"os"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/paketo-buildpacks/packit"

	cmd_detect "github.com/avarteqgmbh/git-ssh-buildpack/cmd/detect"
	"github.com/avarteqgmbh/git-ssh-buildpack/sshagent"
	"github.com/avarteqgmbh/git-ssh-buildpack/utils"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitDetect(t *testing.T) {
	spec.Run(t, "Detect", testDetect, spec.Report(report.Terminal{}))
}

func testDetect(t *testing.T, when spec.G, it spec.S) {
	it.Before(func() {
		RegisterTestingT(t)
	})

	when("the os environment variable is present", func() {
		it("should add git-ssh-buildpack to the buildplan", func() {
			defer utils.ResetEnv(os.Environ())
			os.Clearenv()
			os.Setenv("GIT_SSH_KEY", "VALUE")
			logEmitter := utils.NewLogEmitter(os.Stdout)
			detectFunc := cmd_detect.Detect(logEmitter)
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
							Metadata: cmd_detect.GitSSHBuildPlanMetadata{
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
			logEmitter := utils.NewLogEmitter(os.Stdout)
			detectFunc := cmd_detect.Detect(logEmitter)
			detectResult, err := detectFunc(packit.DetectContext{})
			Expect(err).To(HaveOccurred())
			Expect(detectResult.Plan).To(Equal(packit.BuildPlan{Provides: nil, Requires: nil, Or: nil}))
		})
	})
}
