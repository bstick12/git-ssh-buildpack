package main_test

import (
	"errors"
	"github.com/bstick12/git-key-buildpack/sshagent"
	"io"
	"testing"

	"github.com/buildpack/libbuildpack/buildplan"

	cmdBuild "github.com/bstick12/git-key-buildpack/cmd/build"
	"github.com/cloudfoundry/libcfbuildpack/build"
	"github.com/cloudfoundry/libcfbuildpack/test"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitDetect(t *testing.T) {
	spec.Run(t, "Build", testDetect, spec.Report(report.Terminal{}))
}

type TestRunner struct {
	Runner func () error
}

func(tr *TestRunner) Run() error {
	return tr.Runner()
}

func testDetect(t *testing.T, when spec.G, it spec.S) {

	var factory *test.BuildFactory

	it.Before(func() {
		RegisterTestingT(t)
		factory = test.NewBuildFactory(t)
	})

	when("building source", func() {
		it("should pass if successful", func() {
			sshagent.CmdFunction = func (_, _ io.Writer, _ io.Reader, _ string, _ ...string) sshagent.Runner {
				return &TestRunner {
					Runner: func() error {
						return nil
					},
				}
			}

			factory.Build.BuildPlan = buildplan.BuildPlan{
				sshagent.Dependency: buildplan.Dependency{
					Metadata: buildplan.Metadata{
						"build": true,
					},
				},
			}
			code, err := cmdBuild.RunBuild(factory.Build)
			Expect(err).NotTo(HaveOccurred())
			Expect(code).To(Equal(build.SuccessStatusCode))

			sshagentLayer := factory.Build.Layers.Layer(sshagent.Dependency)
			Expect(sshagentLayer).To(test.HaveLayerMetadata(true, false, false))

		})

		it("should fail if it doesn't contribute", func() {
			sshagent.CmdFunction = func (_, _ io.Writer, _ io.Reader, _ string, _ ...string) sshagent.Runner {
				return &TestRunner {
					Runner: func() error {
						return errors.New("error")
					},
				}
			}
			code, err := cmdBuild.RunBuild(factory.Build)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Failed to find build plan"))
			Expect(code).To(Equal(cmdBuild.FailureStatusCode))
		})
	})

}
