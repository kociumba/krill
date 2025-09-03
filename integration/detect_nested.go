package integration

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/kociumba/krill/config"
)

var excludeDirs = map[string]struct{}{
	".git":                {},
	"node_modules":        {},
	"vendor":              {},
	"includes":            {},
	"target":              {},
	"bin":                 {},
	"build":               {},
	"cmake-build-debug":   {},
	"cmake-build-release": {},
	"meson-build-debug":   {},
	"meson-build-release": {},
}

func DetectNestedProjects(root string) (map[string]config.NestedProject, error) {
	nested := make(map[string]config.NestedProject)

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if _, exclude := excludeDirs[d.Name()]; exclude {
				return filepath.SkipDir
			}
			configPath := filepath.Join(path, "krill.toml")
			if _, err := os.Stat(configPath); err == nil {
				relPath, _ := filepath.Rel(root, path)
				if relPath != "." {
					nested[relPath] = config.NestedProject{Mappings: map[string]string{}}
					// fmt.Printf("Detected nested Krill project at %s\n", relPath)
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return nested, nil
}
