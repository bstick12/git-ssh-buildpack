package main_test

import (
	"os"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/buildpack/libbuildpack/buildplan"

	cmd_detect "github.com/bstick12/git-ssh-buildpack/cmd/detect"
	"github.com/bstick12/git-ssh-buildpack/sshagent"
	"github.com/bstick12/git-ssh-buildpack/utils"

	"github.com/cloudfoundry/libcfbuildpack/detect"
	"github.com/cloudfoundry/libcfbuildpack/test"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitDetect(t *testing.T) {
	spec.Run(t, "Detect", testDetect, spec.Report(report.Terminal{}))
}

func testDetect(t *testing.T, when spec.G, it spec.S) {
	var factory *test.DetectFactory

	it.Before(func() {
		RegisterTestingT(t)
		factory = test.NewDetectFactory(t)
	})

	when("the os environment variable is present", func() {
		it("should add git-ssh-buildpack to the buildplan", func() {
			defer utils.ResetEnv(os.Environ())
			os.Clearenv()
			os.Setenv("GIT_SSH_KEY", "VALUE")
			code, err := cmd_detect.RunDetect(factory.Detect)
			Expect(err).NotTo(HaveOccurred())
			Expect(code).To(Equal(detect.PassStatusCode))
			Expect(factory.Plans.Plan).To(Equal(buildplan.Plan{
				Requires: []buildplan.Required{
					{
						Name:    sshagent.Dependency,
						Version: "",
						Metadata: buildplan.Metadata{
							"build":  true,
							"launch": false,
							"cache":  false,
						},
					},
				},
				Provides: []buildplan.Provided{
					{sshagent.Dependency},
				},
			}))
		})
	})

	when("the os environment variable is not present", func() {
		it("should not add git-ssh-buildpack to the buildplan", func() {
			defer utils.ResetEnv(os.Environ())
			os.Clearenv()
			code, err := cmd_detect.RunDetect(factory.Detect)
			Expect(err).To(HaveOccurred())
			Expect(code).To(Equal(detect.FailStatusCode))
		})
	})
}
