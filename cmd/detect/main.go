package main

import (
	"fmt"
	"os"

	"github.com/bstick12/git-ssh-buildpack/sshagent"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/libcfbuildpack/detect"
	"github.com/pkg/errors"
)

func main() {
	context, err := detect.DefaultDetect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create a default detection context: %s", err)
		os.Exit(detect.FailStatusCode)
	}

	code, err := RunDetect(context)
	if err != nil {
		context.Logger.Info(err.Error())
	}

	os.Exit(code)
}

func RunDetect(context detect.Detect) (int, error) {
	if _, ok := os.LookupEnv("GIT_SSH_KEY"); ok {
		return context.Pass(buildplan.Plan{
			Requires: []buildplan.Required{
				{
					Name: sshagent.Dependency,
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
		})
	}

	return detect.FailStatusCode, errors.New("No GIT_SSH_KEY variable found")
}
