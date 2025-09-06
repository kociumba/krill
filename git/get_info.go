package git

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/kociumba/krill/cli_utils"
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

	var files []string
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		status := strings.TrimSpace(line[:2])
		rest := strings.TrimSpace(line[2:])

		if strings.Contains(rest, " -> ") {
			parts := strings.SplitN(rest, " -> ", 2)
			if len(parts) == 2 {
				files = append(files, fmt.Sprintf("%s %s -> %s", status, parts[0], parts[1]))
				continue
			}
		}

		files = append(files, fmt.Sprintf("%s %s", status, rest))
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

func PrintDetailedGitStatus() {
	if !IsGitRepo() {
		cli_utils.PrintMessage(cli_utils.LevelWarning, "Not a git repository")
		return
	}

	branch, err := CurrentBranch()
	if err != nil {
		branch = "unknown"
	}

	cli_utils.PrintSubHeader(fmt.Sprintf("Git Status - %s", branch), cli_utils.ColorBlue)

	ahead, behind, _ := CommitsAheadBehind()
	if ahead > 0 || behind > 0 {
		syncMsg := ""
		if ahead > 0 {
			syncMsg += fmt.Sprintf("%d ahead", ahead)
		}
		if behind > 0 {
			if syncMsg != "" {
				syncMsg += ", "
			}
			syncMsg += fmt.Sprintf("%d behind", behind)
		}
		cli_utils.PrintMessage(cli_utils.LevelInfo, "Branch "+syncMsg+" of origin")
	} else {
		cli_utils.PrintMessage(cli_utils.LevelSuccess, "Branch up to date with origin")
	}

	fmt.Println()

	conflicts, _ := ConflictedFiles()
	uncommitted, _ := UncommittedFiles()

	if len(conflicts) > 0 {
		cli_utils.PrintMessage(cli_utils.LevelError, fmt.Sprintf("%d conflicted files:", len(conflicts)))
		for _, file := range conflicts {
			cli_utils.PrintIndentedMessage(4, "•", cli_utils.ColorRed, file)
		}

		if len(uncommitted) > 0 {
			fmt.Println()
		}
	}

	if len(uncommitted) > 0 {
		cli_utils.PrintMessage(cli_utils.LevelWarning, fmt.Sprintf("%d unstaged changes:", len(uncommitted)))
		for _, file := range uncommitted {
			cli_utils.PrintIndentedMessage(4, "•", cli_utils.ColorYellow, file)
		}
	}

	if len(uncommitted) == 0 && len(conflicts) == 0 {
		cli_utils.PrintMessage(cli_utils.LevelSuccess, "Working directory clean")
	}
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

func PrintGitStatusCompact() {
	branch, err := CurrentBranch()
	if err != nil {
		branch = "?"
	}

	ahead, behind, _ := CommitsAheadBehind()
	uncommitted, _ := UncommittedFiles()
	conflicts, _ := ConflictedFiles()

	// Branch line with sync status
	branchText := branch
	if ahead > 0 || behind > 0 {
		syncStatus := ""
		if ahead > 0 {
			syncStatus += fmt.Sprintf("↑%d", ahead)
		}
		if behind > 0 {
			syncStatus += fmt.Sprintf("↓%d", behind)
		}
		branchText += " " + syncStatus
	}

	cli_utils.PrintIndentedMessage(2, "→", cli_utils.ColorBlue, branchText)

	// Working directory status
	hasConflicts := len(conflicts) > 0
	hasUncommitted := len(uncommitted) > 0

	switch {
	case hasConflicts:
		cli_utils.PrintIndentedMessage(2, cli_utils.SymbolError, cli_utils.ColorRed,
			fmt.Sprintf("%d conflict(s)", len(conflicts)))
	case hasUncommitted:
		cli_utils.PrintIndentedMessage(2, cli_utils.SymbolWarning, cli_utils.ColorYellow,
			fmt.Sprintf("%d unstaged change(s)", len(uncommitted)))
	default:
		cli_utils.PrintIndentedMessage(2, cli_utils.SymbolSuccess, cli_utils.ColorGreen, "clean")
	}
}
