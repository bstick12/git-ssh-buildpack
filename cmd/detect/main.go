package main

import (
	"fmt"
	"github.com/bstick12/git-key-buildpack/sshagent"
	"os"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/pkg/errors"

	"github.com/cloudfoundry/libcfbuildpack/detect"
)

func main() {
	context, err := detect.DefaultDetect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create a default detection context: %s", err)
		os.Exit(detect.FailStatusCode)
	}

	if err := context.BuildPlan.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize Build Plan: %s\n", err)
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
		return context.Pass(buildplan.BuildPlan{
			sshagent.Dependency : buildplan.Dependency{
				Metadata: buildplan.Metadata{
					"build":  true,
					"launch": false,
				},
			},
		})
	}

	return detect.FailStatusCode, errors.New("No GIT_SSH_KEY variable found")

}
