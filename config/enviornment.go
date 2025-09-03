package config

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func findVsTool(pattern string) (string, error) {
	vswhere := filepath.Join(os.Getenv("ProgramFiles(x86)"),
		"Microsoft Visual Studio", "Installer", "vswhere.exe")
	if _, err := os.Stat(vswhere); err != nil {
		return "", fmt.Errorf("vswhere not found: %w", err)
	}

	cmd := exec.Command(vswhere,
		"-latest",
		"-products", "*",
		"-requires", "Microsoft.VisualStudio.Component.VC.Tools.x86.x64",
		"-find", pattern)

	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", err
	}

	path := strings.TrimSpace(out.String())
	if path == "" {
		return "", fmt.Errorf("%s not found", pattern)
	}
	return path, nil
}

func findVsInstallPath() (string, error) {
	vswhere := filepath.Join(os.Getenv("ProgramFiles(x86)"),
		"Microsoft Visual Studio", "Installer", "vswhere.exe")
	if _, err := os.Stat(vswhere); err != nil {
		return "", fmt.Errorf("vswhere not found: %w", err)
	}

	cmd := exec.Command(vswhere,
		"-latest",
		"-products", "*",
		"-property", "installationPath")

	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", err
	}

	path := strings.TrimSpace(out.String())
	if path == "" {
		return "", fmt.Errorf("installationPath not found")
	}
	return path, nil
}

func DetectEnvironment(requireVsEnv bool) (*Environment, error) {
	if runtime.GOOS == "windows" && requireVsEnv {
		psPath, err := findVsTool("Common7\\Tools\\Launch-VsDevShell.ps1")
		if err == nil {
			installPath, ierr := findVsInstallPath()
			if ierr != nil {
				return nil, ierr
			}

			return &Environment{
				Path: "powershell.exe",
				Args: []string{
					"-NoProfile",
					"-NoLogo",
					"-Command",
					fmt.Sprintf("& { . '%s' -VsInstallPath '%s' -Arch amd64 -HostArch amd64 }", psPath, installPath),
				},
			}, nil
		}

		cmdPath, err := findVsTool("Common7\\Tools\\VsDevCmd.bat")
		if err == nil {
			return &Environment{
				Path: "cmd.exe",
				Args: []string{
					"/c", cmdPath,
				},
			}, nil
		}

		return nil, fmt.Errorf("neither Developer PowerShell nor Developer Cmd found")
	}

	shell := os.Getenv("SHELL")
	if shell == "" {
		switch runtime.GOOS {
		case "windows":
			return &Environment{
				Path: "powershell.exe",
				Args: []string{"-NoProfile", "-NoLogo", "-Command"},
			}, nil
		case "darwin":
			shell = "/bin/zsh"
		default:
			shell = "/bin/bash"
		}
	}

	return &Environment{
		Path: shell,
		Args: []string{"-c"},
	}, nil
}
