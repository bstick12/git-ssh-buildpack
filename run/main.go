package main

import (
	"errors"
	"os"

	sshagent "github.com/avarteqgmbh/git-ssh-buildpack/sshagent"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/scribe"
)

// GitSSHBuildPlanMetadata represents the metadata for this buildpack
type GitSSHBuildPlanMetadata struct {
	Build  bool `toml:"build"`
	Cache  bool `toml:"cache"`
	Launch bool `toml:"launch"`
}

func main() {
	logger := scribe.NewEmitter(os.Stdout).WithLevel(os.Getenv("BP_LOG_LEVEL"))

	packit.Run(
		Detect(),
		Build(logger),
	)
}

// Detect detects whether this Buildpack should participate
func Detect() packit.DetectFunc {
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

func Build(logger scribe.Emitter) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		res, err := sshagent.Contribute(context, logger, sshagent.CmdRunner{})
		return res, err
	}
}
