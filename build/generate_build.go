package build

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"github.com/kociumba/krill/config"
	"github.com/urfave/cli/v3"
)

func GenerateBuildCmds(cfg config.Cfg) []*cli.Command {
	var subcommands []*cli.Command
	for targetName := range cfg.BuildTargets {
		subcommands = append(subcommands, &cli.Command{
			Name:  targetName,
			Usage: fmt.Sprintf("Run build commands for target %s", targetName),
			Action: func(ctx context.Context, cmd *cli.Command) error {
				visited := make(map[string]struct{})
				return buildTarget(ctx, &cfg, targetName, visited)
			},
		})
	}

	return subcommands
}

func buildTarget(ctx context.Context, cfg *config.Cfg, targetName string, visited map[string]struct{}) error {
	target, ok := cfg.BuildTargets[targetName]
	if !ok {
		return fmt.Errorf("Target %s does not exist in the project", targetName)
	}

	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	had_to_detect_env := false

	if _, seen := visited[wd+"-"+targetName]; seen {
		return fmt.Errorf("cycle detected at %s for target %s", wd, targetName)
	}

	visited[wd+"-"+targetName] = struct{}{}

	for _, dep := range target.DependsOn {
		if err := buildTarget(ctx, cfg, dep, visited); err != nil {
			return fmt.Errorf("dependency %s failed: %w", dep, err)
		}
	}

	isAggregate := len(target.DependsOn) > 0 && len(target.Commands) == 0 && target.OutputDir == ""

	isToolSpecific := false
	parts := strings.Split(targetName, "-")
	if len(parts) >= 2 {
		base := parts[0]
		suffix := strings.Join(parts[1:], "-")
		if base == "debug" || base == "release" {
			for _, tool := range cfg.Project.Tools {
				if strings.ToLower(tool.String()) == suffix {
					isToolSpecific = true
					break
				}
			}
		}
	}

	if isAggregate && !isToolSpecific {
		for subPath, subNested := range cfg.Nested {
			subDir := filepath.Join(wd, subPath)
			subCfg, err := config.GetConfigFromDir(subDir)
			if err != nil {
				return fmt.Errorf("failed to load nested config at %s: %w", subPath, err)
			}

			subTarget := targetName
			if mapping, ok := subNested.Mappings[targetName]; ok {
				subTarget = mapping
			}

			if err := os.Chdir(subDir); err != nil {
				return err
			}

			defer os.Chdir(wd)
			if err := buildTarget(ctx, &subCfg, subTarget, visited); err != nil {
				return fmt.Errorf("failed building nested %s: %w", subPath, err)
			}
		}
	}

	if target.OutputDir != "" {
		outputPath := filepath.Join(wd, target.OutputDir)
		if err := os.MkdirAll(outputPath, 0755); err != nil {
			return fmt.Errorf("failed to create output directory %q: %w", target.OutputDir, err)
		}

		gitignorePath := filepath.Join(outputPath, ".gitignore")
		if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
			if err := os.WriteFile(gitignorePath, []byte("*"), 0644); err != nil {
				return fmt.Errorf("failed to write .gitignore in %q: %w", target.OutputDir, err)
			}
		}
	}

	for _, cmd := range target.Commands {
		fmt.Println("Running:", cmd)

		if cfg.Env[runtime.GOOS].Path == "" {
			env, err := config.DetectEnvironment(
				slices.Contains(cfg.Project.Languages, config.C) ||
					slices.Contains(cfg.Project.Languages, config.Cpp))
			if err != nil {
				return err
			}

			had_to_detect_env = true
			cfg.Env[runtime.GOOS] = *env
		}

		args := make([]string, len(cfg.Env[runtime.GOOS].Args))
		copy(args, cfg.Env[runtime.GOOS].Args)

		switch cfg.Env[runtime.GOOS].Path {
		case "powershell.exe":
			if len(args) > 0 && args[len(args)-1] == "-Command" {
				args = append(args, cmd)
			} else {
				args = append(args, cmd)
			}
		case "cmd.exe":
			args = append(args, "&&", cmd)
		default:
			args = append(args, cmd)
		}

		run := exec.CommandContext(ctx, cfg.Env[runtime.GOOS].Path, args...)
		run.Dir = wd
		run.Stderr = os.Stderr
		run.Stdout = os.Stdout
		run.Stdin = os.Stdin

		if err := run.Run(); err != nil {
			return fmt.Errorf("command %q failed: %w", cmd, err)
		}
	}

	if had_to_detect_env {
		reader := bufio.NewReader(os.Stdin)
		fmt.Printf("krill had to detect a default env during this compilation, because '[env.%s]' is not defined in the current config.\n", runtime.GOOS)
		fmt.Print("Do you want to save the detected env to the config? [y/N]: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return err
		}

		input = strings.TrimSpace(input)
		input = strings.ToLower(input)

		switch input {
		case "y", "yes":
			err := config.SaveConfig(config.CFG)
			if err != nil {
				return err
			}
		case "n", "no":
			return nil
		default:
			return nil
		}
	}

	return nil
}

func DefaultTargetsForTool(tool config.Tool, artefact_name string, binType config.BinaryType) map[string]config.BuildTarget {
	targets := make(map[string]config.BuildTarget)

	switch tool {
	case config.CMake:
		targets["debug"] = config.BuildTarget{
			Commands: []string{
				"cmake -S . -B {{ .targets.debug.output_dir }} -DCMAKE_BUILD_TYPE=Debug",
				"cmake --build {{ .targets.debug.output_dir }}",
			},
			OutputDir: "cmake-build-debug",
		}
		targets["release"] = config.BuildTarget{
			Commands: []string{
				"cmake -S . -B {{ .targets.release.output_dir }} -DCMAKE_BUILD_TYPE=Release",
				"cmake --build {{ .targets.release.output_dir }}",
			},
			OutputDir: "cmake-build-release",
		}
	case config.Gradle:
		targets["debug"] = config.BuildTarget{
			Commands:  []string{"./gradlew build -PbuildType=debug"},
			OutputDir: "build",
		}
		targets["release"] = config.BuildTarget{
			Commands:  []string{"./gradlew build -PbuildType=release"},
			OutputDir: "build",
		}
	case config.Meson:
		targets["debug"] = config.BuildTarget{
			Commands: []string{
				"meson setup {{ .targets.debug.output_dir }} --buildtype=debug",
				"meson compile -C {{ .targets.debug.output_dir }}",
			},
			OutputDir: "meson-build-debug",
		}
		targets["release"] = config.BuildTarget{
			Commands: []string{
				"meson setup {{ .targets.release.output_dir }} --buildtype=release",
				"meson compile -C {{ .targets.release.output_dir }}",
			},
			OutputDir: "meson-build-release",
		}
	case config.Cargo:
		targets["debug"] = config.BuildTarget{
			Commands:  []string{"cargo build"},
			OutputDir: "target/debug",
		}
		targets["release"] = config.BuildTarget{
			Commands:  []string{"cargo build --release"},
			OutputDir: "target/release",
		}
	case config.GoCmd:
		targets["debug"] = config.BuildTarget{
			Commands:  []string{"go build -gcflags=\"-N -l\" -o {{ .targets.debug.output_dir }}/{{ .project.name }}{{ .exe_ext }}"},
			OutputDir: "bin/debug",
		}
		targets["release"] = config.BuildTarget{
			Commands:  []string{"go build -ldflags=\"-s -w\" -o {{ .targets.release.output_dir }}/{{ .project.name }}{{ .exe_ext }}"},
			OutputDir: "bin/release",
		}
	case config.OdinCmd:
		targets["debug"] = config.BuildTarget{
			Commands:  []string{"odin build . -debug -out:{{ .targets.debug.output_dir }}/{{ .project.name }}{{ .exe_ext }}"},
			OutputDir: "bin/debug",
		}
		targets["release"] = config.BuildTarget{
			Commands:  []string{"odin build . -o:speed -out:{{ .targets.release.output_dir }}/{{ .project.name }}{{ .exe_ext }}"},
			OutputDir: "bin/release",
		}
	case config.DotNet:
		targets["debug"] = config.BuildTarget{
			Commands:  []string{"dotnet build -c Debug"},
			OutputDir: "bin/Debug",
		}
		targets["release"] = config.BuildTarget{
			Commands:  []string{"dotnet build -c Release"},
			OutputDir: "bin/Release",
		}
	case config.Nob:
		nobBinary := "./nob"
		if runtime.GOOS == "windows" {
			nobBinary = "nob.exe"
		}
		targets["default"] = config.BuildTarget{
			Commands: []string{nobBinary},
		}

	// these defaults might get removed like the raw compiler targets
	case config.Make:
		targets["debug"] = config.BuildTarget{
			Commands: []string{"make debug"},
		}
		targets["release"] = config.BuildTarget{
			Commands: []string{"make release"},
		}
	case config.Taskfile:
		targets["debug"] = config.BuildTarget{
			Commands: []string{"task build:debug"},
		}
		targets["release"] = config.BuildTarget{
			Commands: []string{"task build:release"},
		}
	}

	return targets
}

func GenerateDefaultBuildTargets(cfg *config.Cfg) error {
	if len(cfg.Project.Tools) == 0 {
		return fmt.Errorf("no tools detected to generate build targets")
	}

	artefact_name := cfg.Project.Name
	bin_type := config.Executable
	if cfg.Project.BinaryType != 0 {
		bin_type = cfg.Project.BinaryType
	}

	artefact_name += config.BinaryTypeToExt[bin_type]

	isMulti := len(cfg.Project.Tools) > 1
	if !isMulti && len(cfg.Project.Languages) > 1 {
		tool := cfg.Project.Tools[0]
		supportedLangs := config.ToolToLang[tool]
		supported := true
		for _, lang := range cfg.Project.Languages {
			if _, ok := supportedLangs[lang]; !ok {
				supported = false
				break
			}
		}

		if !supported {
			isMulti = true
		}
	}

	cfg.BuildTargets = make(map[string]config.BuildTarget)

	var debugDeps, releaseDeps []string

	for _, tool := range cfg.Project.Tools {
		toolTargets := DefaultTargetsForTool(tool, artefact_name, bin_type)
		for baseName, tgt := range toolTargets {
			newName := baseName
			if isMulti {
				newName = fmt.Sprintf("%s-%s", baseName, strings.ToLower(tool.String()))
			}

			cfg.BuildTargets[newName] = tgt

			if isMulti {
				if strings.HasPrefix(baseName, "debug") {
					debugDeps = append(debugDeps, newName)
				} else if strings.HasPrefix(baseName, "release") {
					releaseDeps = append(releaseDeps, newName)
				}
			}
		}
	}

	if isMulti {
		if len(debugDeps) > 0 {
			cfg.BuildTargets["debug"] = config.BuildTarget{DependsOn: debugDeps}
		}
		if len(releaseDeps) > 0 {
			cfg.BuildTargets["release"] = config.BuildTarget{DependsOn: releaseDeps}
		}
	}

	return nil
}
