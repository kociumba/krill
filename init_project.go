package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"

	"github.com/kociumba/krill/build"
	"github.com/kociumba/krill/config"
	"github.com/kociumba/krill/integration"
)

func InitProject() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	config.CFG.Project.BinaryType = config.Executable

	name := filepath.Base(wd)
	config.CFG.Project.Name = name

	ver, err := config.GetVersion()
	if err != nil {
		fmt.Printf("Could not determine the project version using default '%s'\n", ver.StringV())
	}

	config.CFG.Project.Version = ver.String()

	tools := config.DetectTools(wd)
	if len(tools) == 0 {
		fmt.Println("No tools supported by krill have been found in this directory")
	}

	config.CFG.Project.Tools = tools

	langs := config.DetectLanguages(wd, tools)
	if len(langs) == 0 {
		fmt.Println("No languages supported by krill have been found in this directory")
	}

	config.CFG.Project.Languages = langs

	env, err := config.DetectEnvironment(slices.Contains(langs, config.C) || slices.Contains(langs, config.Cpp))
	config.CFG.Env[runtime.GOOS] = *env

	err = build.GenerateDefaultBuildTargets(&config.CFG)
	if err != nil {
		return err
	}

	nested, err := integration.DetectNestedProjects(wd)
	if err != nil {
		return err
	}

	config.CFG.Nested = nested

	err = config.SaveConfig(config.CFG)
	if err != nil {
		return err
	}

	config.CFG, err = config.GetConfig()
	if err != nil {
		return err
	}

	return nil
}
