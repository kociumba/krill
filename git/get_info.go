package git

import (
	"fmt"
	"os/exec"
	"strings"
)

func CheckGitInstalled() bool {
	cmd := exec.Command("git", "--version")
	if err := cmd.Run(); err != nil {
		return false
	}

	return true
}

func IsGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	cmd.Stderr = nil
	out, err := cmd.Output()
	if err != nil {
		return false
	}

	return strings.TrimSpace(string(out)) == "true"
}

func CurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(out)), nil
}

func CommitsAheadBehind() (ahead int, behind int, err error) {
	cmd := exec.Command("git", "remote")
	out, err := cmd.Output()
	if err != nil || !strings.Contains(string(out), "origin") {
		return 0, 0, nil
	}

	cmd = exec.Command("git", "rev-list", "--left-right", "--count", "HEAD...@{u}")
	out, err = cmd.Output()
	if err != nil {
		return 0, 0, err
	}

	var aheadCount, behindCount int
	_, err = fmt.Sscanf(string(out), "%d\t%d", &aheadCount, &behindCount)
	if err != nil {
		return 0, 0, err
	}

	return aheadCount, behindCount, nil
}

func UncommittedFiles() ([]string, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var files []string
	for _, line := range lines {
		if line == "" {
			continue
		}

		status := line[:2]
		file := strings.TrimSpace(line[3:])
		files = append(files, fmt.Sprintf("%s %s", status, file))
	}

	return files, nil
}

func ConflictedFiles() ([]string, error) {
	cmd := exec.Command("git", "diff", "--name-only", "--diff-filter=U")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var files []string
	for _, line := range lines {
		if line != "" {
			files = append(files, line)
		}
	}

	return files, nil
}

func FormatGitStatusString() string {
	if !IsGitRepo() {
		return ""
	}

	branch, err := CurrentBranch()
	if err != nil {
		branch = "?"
	}

	ahead, behind, _ := CommitsAheadBehind()

	uncommitted, _ := UncommittedFiles()
	conflicts, _ := ConflictedFiles()

	var b strings.Builder

	const indent = "  "

	branchStatus := branch
	if ahead > 0 || behind > 0 {
		branchStatus += " "
		if ahead > 0 {
			branchStatus += fmt.Sprintf("↑%d", ahead)
		}
		if behind > 0 {
			branchStatus += fmt.Sprintf("↓%d", behind)
		}
	}

	fmt.Fprintf(&b, " %s\n", branchStatus)

	fmt.Fprintln(&b)

	hasUncommitted := len(uncommitted) > 0
	hasConflicts := len(conflicts) > 0

	switch {
	case hasUncommitted && hasConflicts:
		fmt.Fprintf(&b, "%sUnstaged:\n", indent)
		for _, file := range uncommitted {
			fmt.Fprintf(&b, "%s  %s\n", indent, file)
		}

		fmt.Fprintln(&b)
		fmt.Fprintf(&b, "%sConflicts:\n", indent)
		for _, file := range conflicts {
			fmt.Fprintf(&b, "%s  %s\n", indent, file)
		}

	case hasUncommitted:
		fmt.Fprintf(&b, "%sUnstaged:\n", indent)
		for _, file := range uncommitted {
			fmt.Fprintf(&b, "%s  %s\n", indent, file)
		}

	case hasConflicts:
		fmt.Fprintf(&b, "%sConflicts:\n", indent)
		for _, file := range conflicts {
			fmt.Fprintf(&b, "%s  %s\n", indent, file)
		}

	default:
		fmt.Fprintf(&b, "%sClean ✓\n", indent)
	}

	return b.String()
}
