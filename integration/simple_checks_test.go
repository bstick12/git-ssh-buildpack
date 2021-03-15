package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/occam/matchers"
)

func testSimpleChecks(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		pack   occam.Pack
		docker occam.Docker
	)

	it.Before(func() {
		pack = occam.NewPack()
		docker = occam.NewDocker()
	})

	context("run build with and without environment variables", func() {
		var (
			image occam.Image

			name       string
			source     string
			keyfile    []byte
			userSSHKey string
			sshKey     string
			ok         bool
		)

		it.Before(func() {
			var err error
			name, err = occam.RandomName()
			Expect(err).NotTo(HaveOccurred())

			keyfile, err = os.ReadFile("./testdata/.ssh/id_rsa")
			Expect(err).NotTo(HaveOccurred())

			sshKey = string(keyfile)
			Expect(err).NotTo(HaveOccurred())

			userSSHKey, ok = os.LookupEnv("GIT_SSH_KEY")
			if ok {
				defer os.Setenv("GIT_SSH_KEY", userSSHKey)
				os.Unsetenv("GIT_SSH_KEY")
			}
		})

		it.After(func() {
			Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
			Expect(os.RemoveAll(source)).To(Succeed())
		})

		it("fails without defined environment variables", func() {
			var err error
			source, err = occam.Source(filepath.Join("testdata", "default_app"))
			Expect(err).NotTo(HaveOccurred())

			var logs fmt.Stringer
			image, logs, err = pack.WithNoColor().Build.
				WithPullPolicy("never").
				WithBuildpacks(
					settings.Buildpacks.Sshagent.Online,
					settings.Buildpacks.BuildPlan.Online,
				).
				Execute(name, source)
			Expect(err).To(HaveOccurred(), logs.String)

			Expect(logs).To(ContainLines(
				"No GIT_SSH_KEY variable found",
			))
		})

		it("installs with defined environment variables", func() {
			var err error
			source, err = occam.Source(filepath.Join("testdata", "default_app"))
			Expect(err).NotTo(HaveOccurred())

			var logs fmt.Stringer
			image, logs, err = pack.WithNoColor().Build.
				WithPullPolicy("never").
				WithBuildpacks(
					settings.Buildpacks.Sshagent.Online,
					settings.Buildpacks.BuildPlan.Online,
				).
				WithEnv(map[string]string{
					"GIT_SSH_KEY":          sshKey,
					"GIT_SSH_DONT_CONNECT": "TRUE",
				}).
				Execute(name, source)
			Expect(err).ToNot(HaveOccurred(), logs.String)

			Expect(logs).To(ContainLines(
				MatchRegexp(fmt.Sprintf(`%s \d+\.\d+\.\d+`, settings.Buildpack.Name)),
				"    Starting SSH agent",
				ContainSubstring("Identity added: (stdin)"),
				"    Configuring host [github.com]",
			))
			Expect(docker.Image.Remove.Execute(image.ID)).To(Succeed())
		})
	})
}
