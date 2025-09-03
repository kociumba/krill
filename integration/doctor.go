package integration

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/kociumba/krill/build"
	"github.com/kociumba/krill/config"
	"github.com/kociumba/krill/git"
)

const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorCyan   = "\033[36m"
	ColorGray   = "\033[37m"

	SymbolError   = "✗"
	SymbolWarning = "⚠"
	SymbolSuccess = "✓"
	SymbolInfo    = "ℹ"
	SymbolFix     = "→"
)

type Issue struct {
	Category    string
	Level       string
	Description string
	Fix         string
}

func Doctor(save_changes, show_diff bool) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	var detected_langs []config.Language
	var issues []Issue
	var fixedCfg config.Cfg

	if !git.CheckGitInstalled() {
		issues = append(issues, Issue{
			Category:    "Global",
			Level:       "error",
			Description: "Git is not installed or not found in PATH",
			Fix:         "Install git or add it to your PATH environment variable",
		})
	}

	if !config.HasConfig {
		issues = append(issues, Issue{
			Category:    "Configuration",
			Level:       "warning",
			Description: "No project config found in working directory",
			Fix:         "Run 'krill init' to create a project configuration",
		})
	} else {
		fixedCfg = config.CFG

		detected_tools := config.DetectTools(wd)
		isMulti := len(detected_tools) > 1 || len(config.DetectLanguages(wd, detected_tools)) > 1
		wasMulti := len(config.CFG.Project.Tools) > 1 || len(config.CFG.Project.Languages) > 1

		if !config.EqualTools(config.CFG.Project.Tools, detected_tools) || isMulti != wasMulti {
			issues = append(issues, Issue{
				Category:    "Configuration",
				Level:       "warning",
				Description: "Tooling configuration differs from detected setup",
				Fix:         "Update configuration to match detected tools",
			})
			fixedCfg.Project.Tools = detected_tools
			fixedCfg.Project.Languages = config.DetectLanguages(wd, detected_tools)

			if err := build.GenerateDefaultBuildTargets(&fixedCfg); err != nil {
				issues = append(issues, Issue{
					Category:    "Build",
					Level:       "error",
					Description: fmt.Sprintf("Failed to regenerate build targets: %v", err),
				})
			} else {
				if isMulti {
					issues = append(issues, Issue{
						Category:    "Build",
						Level:       "info",
						Description: "Multi-language project detected",
						Fix:         "Added prefixed and aggregate build targets",
					})
				}
			}
		} else {
			detected_langs = config.DetectLanguages(wd, detected_tools)
			if !config.EqualLanguages(config.CFG.Project.Languages, detected_langs) {
				issues = append(issues, Issue{
					Category:    "Configuration",
					Level:       "warning",
					Description: "Language configuration differs from detected languages",
					Fix:         "Update configuration to match detected languages",
				})
				fixedCfg.Project.Languages = detected_langs
			}
		}

		nested, err := DetectNestedProjects(wd)
		if err == nil && !config.EqualNested(config.CFG.Nested, nested) {
			issues = append(issues, Issue{
				Category:    "Structure",
				Level:       "warning",
				Description: "Nested project configuration mismatch",
				Fix:         "Update nested project settings",
			})
			fixedCfg.Nested = nested
		}
	}

	printDoctorResults(issues)

	if len(issues) > 0 {
		if show_diff {
			printConfigDiff(config.CFG, fixedCfg)
		}

		if save_changes {
			if err := applyFixes(fixedCfg); err != nil {
				return err
			}
		} else if hasFixableIssues(issues) {
			printFixInstructions(save_changes, show_diff)
		}
	}

	return nil
}

func printDoctorResults(issues []Issue) {
	fmt.Printf("\n%s%s Krill Doctor Results%s\n", ColorCyan, SymbolInfo, ColorReset)
	fmt.Printf("%s%s%s\n", ColorGray, strings.Repeat("─", 50), ColorReset)

	if len(issues) == 0 {
		fmt.Printf("\n%s%s No issues detected - your project looks healthy!%s\n\n",
			ColorGreen, SymbolSuccess, ColorReset)
		return
	}

	categories := make(map[string][]Issue)
	for _, issue := range issues {
		categories[issue.Category] = append(categories[issue.Category], issue)
	}

	for category, categoryIssues := range categories {
		fmt.Printf("\n%s%s %s Issues:%s\n", ColorBlue, SymbolInfo, category, ColorReset)

		for _, issue := range categoryIssues {
			symbol, color := getIssueFormatting(issue.Level)
			fmt.Printf("  %s%s%s %s\n", color, symbol, ColorReset, issue.Description)

			if issue.Fix != "" {
				fmt.Printf("    %s%s%s %s\n", ColorGray, SymbolFix, ColorReset, issue.Fix)
			}
		}
	}

	errorCount := countIssuesByLevel(issues, "error")
	warningCount := countIssuesByLevel(issues, "warning")
	infoCount := countIssuesByLevel(issues, "info")

	fmt.Printf("\n%sSummary:%s\n", ColorCyan, ColorReset)
	if errorCount > 0 {
		fmt.Printf("  %s%d error(s)%s\n", ColorRed, errorCount, ColorReset)
	}

	if warningCount > 0 {
		fmt.Printf("  %s%d warning(s)%s\n", ColorYellow, warningCount, ColorReset)
	}

	if infoCount > 0 {
		fmt.Printf("  %s%d info message(s)%s\n", ColorBlue, infoCount, ColorReset)
	}

	fmt.Println()
}

func printConfigDiff(current, fixed config.Cfg) {
	fmt.Printf("\n%s%s Configuration Changes:%s\n", ColorCyan, SymbolInfo, ColorReset)
	fmt.Printf("%s%s%s\n", ColorGray, strings.Repeat("─", 30), ColorReset)

	curBytes, _ := toml.Marshal(current)
	fixedBytes, _ := toml.Marshal(fixed)

	printLineDiff(string(curBytes), string(fixedBytes))
	fmt.Println()
}

func printLineDiff(a, b string) {
	la := strings.Split(a, "\n")
	lb := strings.Split(b, "\n")
	seen := make(map[string]struct{})

	for _, line := range la {
		seen[line] = struct{}{}
	}

	for _, line := range la {
		if !slices.Contains(lb, line) && strings.TrimSpace(line) != "" {
			fmt.Printf("  %s-%s %s\n", ColorRed, ColorReset, line)
		}
	}

	for _, line := range lb {
		if _, ok := seen[line]; !ok && strings.TrimSpace(line) != "" {
			fmt.Printf("  %s+%s %s\n", ColorGreen, ColorReset, line)
		}
	}
}

func applyFixes(fixedCfg config.Cfg) error {
	fmt.Printf("%s%s Applying configuration fixes...%s\n", ColorYellow, SymbolInfo, ColorReset)

	if err := config.SaveConfig(fixedCfg); err != nil {
		fmt.Printf("%s%s Failed to save configuration: %v%s\n", ColorRed, SymbolError, err, ColorReset)
		return err
	}

	fmt.Printf("%s%s Configuration updated successfully!%s\n\n", ColorGreen, SymbolSuccess, ColorReset)
	return nil
}

func printFixInstructions(save_changes, show_diff bool) {
	if !save_changes {
		fmt.Printf("%s%s To apply these fixes automatically, run:%s\n", ColorYellow, SymbolInfo, ColorReset)
		fmt.Printf("  %skrill doctor -auto-fix%s\n\n", ColorGray, ColorReset)
	}

	if !show_diff {
		fmt.Printf("%s%s To see proposed changes before applying, run:%s\n", ColorYellow, SymbolInfo, ColorReset)
		fmt.Printf("  %skrill doctor -diff%s\n\n", ColorGray, ColorReset)
	}
}

func getIssueFormatting(level string) (string, string) {
	switch level {
	case "error":
		return SymbolError, ColorRed
	case "warning":
		return SymbolWarning, ColorYellow
	case "info":
		return SymbolInfo, ColorBlue
	default:
		return SymbolInfo, ColorGray
	}
}

func countIssuesByLevel(issues []Issue, level string) int {
	count := 0
	for _, issue := range issues {
		if issue.Level == level {
			count++
		}
	}

	return count
}

func hasFixableIssues(issues []Issue) bool {
	for _, issue := range issues {
		if issue.Level == "warning" || issue.Level == "error" {
			return true
		}
	}

	return false
}
