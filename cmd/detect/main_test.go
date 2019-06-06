package main_test

import (
	"github.com/bstick12/git-key-buildpack/sshagent"
	"os"
	"testing"

	"github.com/buildpack/libbuildpack/buildplan"

	cmd_detect "github.com/bstick12/git-key-buildpack/cmd/detect"
	"github.com/bstick12/git-key-buildpack/utils"
	"github.com/cloudfoundry/libcfbuildpack/detect"
	"github.com/cloudfoundry/libcfbuildpack/test"
	. "github.com/onsi/gomega"
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
		it("should add git-key-buildpack to the buildplan", func() {
			defer utils.ResetEnv(os.Environ())
			os.Clearenv()
			os.Setenv("GIT_SSH_KEY", "VALUE")
			code, err := cmd_detect.RunDetect(factory.Detect)
			Expect(err).NotTo(HaveOccurred())
			Expect(code).To(Equal(detect.PassStatusCode))
			Expect(factory.Output).To(Equal(buildplan.BuildPlan{
				sshagent.Dependency: buildplan.Dependency{
					Metadata: buildplan.Metadata{
						"build":  true,
						"launch": false,
					},
				},
			}))
		})
	})

	when("the os environment variable is not present", func() {
		it("should not add git-key-buildpack to the buildplan", func() {
			defer utils.ResetEnv(os.Environ())
			os.Clearenv()
			code, err := cmd_detect.RunDetect(factory.Detect)
			Expect(err).To(HaveOccurred())
			Expect(code).To(Equal(detect.FailStatusCode))
		})
	})

}
