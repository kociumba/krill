package config

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

var default_version = Version{
	0, 0, 0, "",
}

type Version struct {
	Major           int
	Minor           int
	Patch           int
	Version_postfix string // for versions like v0.1.4-p1
}

func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d%s", v.Major, v.Minor, v.Patch, v.Version_postfix)
}

func (v Version) StringV() string {
	return fmt.Sprintf("v%d.%d.%d%s", v.Major, v.Minor, v.Patch, v.Version_postfix)
}

func ParseVersion(s string) (Version, error) {
	re := regexp.MustCompile(`^v?(\d+)\.(\d+)(?:\.(\d+))?([-+].*)?$`)
	matches := re.FindStringSubmatch(strings.TrimSpace(s))
	if len(matches) == 0 {
		return Version{}, fmt.Errorf("invalid version format: %s", s)
	}

	major, err := strconv.Atoi(matches[1])
	if err != nil {
		return Version{}, err
	}

	minor, err := strconv.Atoi(matches[2])
	if err != nil {
		return Version{}, err
	}

	patch := 0
	if matches[3] != "" {
		patch, err = strconv.Atoi(matches[3])
		if err != nil {
			return Version{}, err
		}
	}

	postfix := matches[4]

	return Version{
		Major:           major,
		Minor:           minor,
		Patch:           patch,
		Version_postfix: postfix,
	}, nil
}

func GetVersion() (Version, error) {
	if HasConfig {
		return ParseVersion(CFG.Project.Version)
	}

	if err := exec.Command("git", "rev-parse", "--is-inside-work-tree").Run(); err != nil {
		return default_version, fmt.Errorf("no config and not in a git repository: %w", err)
	}

	cmd := exec.Command("git", "describe", "--tags")
	out, err := cmd.Output()
	if err != nil {
		return ParseVersion("v0.0.0")
	}

	tag := strings.TrimSpace(string(out))

	return ParseVersion(tag)
}
