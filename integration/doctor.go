package integration

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/kociumba/krill/build"
	"github.com/kociumba/krill/cli_utils"
	"github.com/kociumba/krill/config"
	"github.com/kociumba/krill/git"
)

type Issue struct {
	Category    string
	Level       cli_utils.MessageLevel
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
			Level:       cli_utils.LevelError,
			Description: "Git is not installed or not found in PATH",
			Fix:         "Install git or add it to your PATH environment variable",
		})
	}

	if !config.HasConfig {
		issues = append(issues, Issue{
			Category:    "Configuration",
			Level:       cli_utils.LevelWarning,
			Description: "No project config found in working directory",
			Fix:         "Run 'krill init' to create a project configuration",
		})
	} else {
		fixedCfg = config.CFG_unexpanded

		detected_tools := config.DetectTools(wd)
		isMulti := len(detected_tools) > 1 || len(config.DetectLanguages(wd, detected_tools)) > 1
		wasMulti := len(config.CFG_unexpanded.Project.Tools) > 1 || len(config.CFG_unexpanded.Project.Languages) > 1

		if !config.EqualTools(config.CFG_unexpanded.Project.Tools, detected_tools) || isMulti != wasMulti {
			issues = append(issues, Issue{
				Category:    "Configuration",
				Level:       cli_utils.LevelWarning,
				Description: "Tooling configuration differs from detected setup",
				Fix:         "Update configuration to match detected tools",
			})
			fixedCfg.Project.Tools = detected_tools
			fixedCfg.Project.Languages = config.DetectLanguages(wd, detected_tools)

			if err := build.GenerateDefaultBuildTargets(&fixedCfg); err != nil {
				issues = append(issues, Issue{
					Category:    "Build",
					Level:       cli_utils.LevelError,
					Description: fmt.Sprintf("Failed to regenerate build targets: %v", err),
				})
			} else {
				if isMulti {
					issues = append(issues, Issue{
						Category:    "Build",
						Level:       cli_utils.LevelInfo,
						Description: "Multi-language project detected",
						Fix:         "Added prefixed and aggregate build targets",
					})
				}
			}
		} else {
			detected_langs = config.DetectLanguages(wd, detected_tools)
			if !config.EqualLanguages(config.CFG_unexpanded.Project.Languages, detected_langs) {
				issues = append(issues, Issue{
					Category:    "Configuration",
					Level:       cli_utils.LevelWarning,
					Description: "Language configuration differs from detected languages",
					Fix:         "Update configuration to match detected languages",
				})
				fixedCfg.Project.Languages = detected_langs
			}
		}

		nested, err := DetectNestedProjects(wd)
		if err == nil && !config.EqualNested(config.CFG_unexpanded.Nested, nested) {
			issues = append(issues, Issue{
				Category:    "Structure",
				Level:       cli_utils.LevelWarning,
				Description: "Nested project configuration mismatch",
				Fix:         "Update nested project settings",
			})
			fixedCfg.Nested = nested
		}
	}

	displayDoctorResults(issues)

	if len(issues) > 0 {
		if show_diff {
			displayConfigDiff(config.CFG_unexpanded, fixedCfg)
		}

		if save_changes {
			if err := applyConfigurationFixes(fixedCfg); err != nil {
				return err
			}
		} else if hasFixableIssues(issues) {
			displayFixInstructions(save_changes, show_diff)
		}
	}

	return nil
}

func displayDoctorResults(issues []Issue) {
	cli_utils.PrintHeader("Krill Doctor Results", cli_utils.ColorCyan)

	if len(issues) == 0 {
		cli_utils.PrintNoIssuesFound("your project")
		return
	}

	categories := groupIssuesByCategory(issues)

	for category, categoryIssues := range categories {
		cli_utils.PrintSubHeader(category+" Issues", cli_utils.ColorBlue)

		for _, issue := range categoryIssues {
			cli_utils.PrintMessage(issue.Level, issue.Description)
			if issue.Fix != "" {
				cli_utils.PrintIndentedMessage(4, cli_utils.SymbolFix, cli_utils.ColorGray, issue.Fix)
			}
		}
	}

	summary := createIssueSummary(issues)
	cli_utils.PrintSummary(summary)
}

func displayConfigDiff(current, fixed config.Cfg) {
	curBytes, _ := toml.Marshal(current)
	fixedBytes, _ := toml.Marshal(fixed)

	cli_utils.PrintDiff("Configuration Changes", string(curBytes), string(fixedBytes))
}

func applyConfigurationFixes(fixedCfg config.Cfg) error {
	cli_utils.PrintProgressMessage("Applying configuration fixes")

	if err := config.SaveConfig(fixedCfg); err != nil {
		cli_utils.PrintErrorMessage(fmt.Sprintf("Failed to save configuration: %v", err))
		return err
	}

	cli_utils.PrintCompletionMessage("Configuration updated successfully")
	return nil
}

func displayFixInstructions(save_changes, show_diff bool) {
	cli_utils.PrintInstructions(
		"To apply these fixes automatically, run",
		"krill doctor --auto-fix",
		save_changes,
	)

	cli_utils.PrintInstructions(
		"To see proposed changes before applying, run",
		"krill doctor --diff",
		show_diff,
	)
}

func groupIssuesByCategory(issues []Issue) map[string][]Issue {
	categories := make(map[string][]Issue)
	for _, issue := range issues {
		categories[issue.Category] = append(categories[issue.Category], issue)
	}

	return categories
}

func createIssueSummary(issues []Issue) cli_utils.CountSummary {
	var summary cli_utils.CountSummary

	for _, issue := range issues {
		switch issue.Level {
		case cli_utils.LevelError:
			summary.Errors++
		case cli_utils.LevelWarning:
			summary.Warnings++
		case cli_utils.LevelInfo:
			summary.Info++
		case cli_utils.LevelSuccess:
			summary.Success++
		}
	}

	return summary
}

func hasFixableIssues(issues []Issue) bool {
	for _, issue := range issues {
		if issue.Level == cli_utils.LevelWarning || issue.Level == cli_utils.LevelError {
			return true
		}
	}

	return false
}
