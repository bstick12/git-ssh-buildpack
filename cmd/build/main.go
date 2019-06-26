package main

import (
	"fmt"
	"github.com/bstick12/git-ssh-buildpack/sshagent"
	"os"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/pkg/errors"

	"github.com/cloudfoundry/libcfbuildpack/build"
)

const (
	FailureStatusCode = 103
)

func main() {

	context, err := build.DefaultBuild()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to create a default build context: %s", err)
		os.Exit(101)
	}

	code, err := RunBuild(context, sshagent.CmdRunner{})
	if err != nil {
		context.Logger.Info(err.Error())
	}

	os.Exit(code)

}

func RunBuild(context build.Build, runner sshagent.Runner) (int, error) {
	context.Logger.FirstLine(context.Logger.PrettyIdentity(context.Buildpack))

	err := sshagent.Contribute(context, runner)
	if err != nil {
		return context.Failure(FailureStatusCode), errors.Errorf("Failed to find build plan to create Contributor for %s - [%v]", "ssh-agent", err)

	}
	return context.Success(buildplan.BuildPlan{})


}
