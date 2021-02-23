package main

import (
	"os"

	"github.com/avarteqgmbh/git-ssh-buildpack/sshagent"
	"github.com/avarteqgmbh/git-ssh-buildpack/utils"

	"github.com/paketo-buildpacks/packit"
	"github.com/pkg/errors"
)

// GitSSHBuildPlanMetadata represents the metadata for this buildpack
type GitSSHBuildPlanMetadata struct {
	Build  bool `toml:"build"`
	Cache  bool `toml:"cache"`
	Launch bool `toml:"launch"`
}

func main() {
	logEmitter := utils.NewLogEmitter(os.Stdout)
	packit.Detect(Detect(logEmitter))
}

// Detect detects whether this Buildpack should participate
func Detect(logger utils.LogEmitter) packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {
		if _, ok := os.LookupEnv("GIT_SSH_KEY"); ok {
			return packit.DetectResult{
				Plan: packit.BuildPlan{
					Provides: []packit.BuildPlanProvision{
						{Name: sshagent.Dependency},
					},
					Requires: []packit.BuildPlanRequirement{
						{
							Name: sshagent.Dependency,
							Metadata: GitSSHBuildPlanMetadata{
								Build:  true,
								Cache:  false,
								Launch: false,
							},
						},
					},
				},
			}, nil
		}

		return packit.DetectResult{}, errors.New("No GIT_SSH_KEY variable found")
	}
}
