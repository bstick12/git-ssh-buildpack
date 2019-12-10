package utils_test

import (
	"os"
	"testing"

	"github.com/avarteqgmbh/git-ssh-buildpack/utils"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitDetect(t *testing.T) {
	spec.Run(t, "Utils", testDetect, spec.Report(report.Terminal{}))
}

func testDetect(t *testing.T, when spec.G, it spec.S) {

	it.Before(func() {
		RegisterTestingT(t)
	})

	when("reseting environment", func() {
		it("has the same value prior to reset", func() {
			envVars := os.Environ()
			os.Clearenv()
			Expect(os.Environ()).To(HaveLen(0))
			utils.ResetEnv(envVars)
			Expect(os.Environ()).To(Equal(envVars))
		})
	})
}
