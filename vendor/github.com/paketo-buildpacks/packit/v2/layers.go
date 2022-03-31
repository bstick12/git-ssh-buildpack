package packit

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Layers represents the set of layers managed by a buildpack.
type Layers struct {
	// Path is the absolute location of the set of layers managed by a buildpack
	// on disk.
	Path string
}

// Get will either create a new layer with the given name and layer types. If a
// layer already exists on disk, then the layer metadata will be retrieved from
// disk and returned instead.
func (l Layers) Get(name string) (Layer, error) {
	layer := Layer{
		Path:             filepath.Join(l.Path, name),
		Name:             name,
		SharedEnv:        Environment{},
		BuildEnv:         Environment{},
		LaunchEnv:        Environment{},
		ProcessLaunchEnv: make(map[string]Environment),
	}

	_, err := toml.DecodeFile(filepath.Join(l.Path, fmt.Sprintf("%s.toml", name)), &layer)
	if err != nil {
		if !os.IsNotExist(err) {
			return Layer{}, fmt.Errorf("failed to parse layer content metadata: %s", err)
		}
	}

	layer.SharedEnv, err = newEnvironmentFromPath(filepath.Join(l.Path, name, "env"))
	if err != nil {
		return Layer{}, err
	}

	layer.BuildEnv, err = newEnvironmentFromPath(filepath.Join(l.Path, name, "env.build"))
	if err != nil {
		return Layer{}, err
	}

	layer.LaunchEnv, err = newEnvironmentFromPath(filepath.Join(l.Path, name, "env.launch"))
	if err != nil {
		return Layer{}, err
	}

	if _, err := os.Stat(filepath.Join(l.Path, name, "env.launch")); !os.IsNotExist(err) {
		paths, err := os.ReadDir(filepath.Join(l.Path, name, "env.launch"))
		if err != nil {
			return Layer{}, err
		}

		for _, path := range paths {
			if path.IsDir() {
				layer.ProcessLaunchEnv[path.Name()], err = newEnvironmentFromPath(filepath.Join(l.Path, name, "env.launch", path.Name()))
				if err != nil {
					return Layer{}, err
				}
			}
		}
	}

	return layer, nil
}
