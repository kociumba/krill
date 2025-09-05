package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/kociumba/krill/build"
	"github.com/kociumba/krill/cli_utils"
	"github.com/kociumba/krill/config"
	"github.com/kociumba/krill/git"
	"github.com/kociumba/krill/integration"
	"github.com/kociumba/krill/templating"
	"github.com/urfave/cli/v3"
)

var build_cmds []*cli.Command
var build_action cli.ActionFunc

// all of the unimplemented features are conceptual and might not be added
var cmds = []*cli.Command{
	{
		Name:  "init",
		Usage: "Initialize a new project (create config, detect build system, etc.)",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if config.HasConfig {
				ok, err := cli_utils.Prompt("The current directory already has an initialized krill project.\nReinitialize?")
				if err != nil {
					return err
				}

				if ok {
					os.Remove("krill.toml")
				} else {
					if cli_utils.SkipNO {
						fmt.Println("skipped reinitializing the project config since it already exists and promps are skipped using 'no'")
					}

					if cli_utils.SkipYES {
						fmt.Println("reinitialized the project config, since prompts are skipped using 'yes'")
					}

					return nil
				}
			}

			err = InitProject()
			if err != nil {
				return err
			}

			fmt.Printf("Initialized project '%s'\n", config.CFG.Project.Name)
			// fmt.Printf("\nEdit 'krill.toml' to manage this project\n")
			return nil
		},
	},
	{
		Name:  "doctor",
		Usage: "Detects issues in the current config, and missing tools like git",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return integration.Doctor(cmd.Bool("auto-fix"), cmd.Bool("diff"))
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "auto-fix",
				Usage: "Allow the doctor command to automatically apply generated fixes to a config",
			},
			&cli.BoolFlag{
				Name:  "diff",
				Usage: "Show the diff that would be applied by krill if using --auto-fix",
			},
		},
	},
	{
		Name:     "run",
		Usage:    "run targets defined in the config file",
		HideHelp: true,
	},
	{
		Name:  "status",
		Usage: "Show a quick overview of the status of the project",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if config.HasConfig {
				ver, err := config.GetVersion()
				if err != nil {
					return err
				}

				fmt.Print(config.CFG.Project.Name + " " + ver.StringV())
			} else {
				wd, err := os.Getwd()
				if err != nil {
					return err
				}

				name := filepath.Base(wd)
				ver, err := config.GetVersion()
				if err != nil {
					return err
				}

				fmt.Print(name + " " + ver.StringV())
			}

			git_status := git.FormatGitStatusString()

			if git_status != "" {
				fmt.Println("", git_status)
			}

			if config.HasConfig {
				fmt.Println("krill config ✓")
			} else {
				fmt.Println("no krill config, build features unavailible")
			}

			return nil
		},
	},
	{
		Name:  "debug",
		Usage: "All debugging utilities are groupped under this subcommand",
		Commands: []*cli.Command{
			{
				Name:  "expand-cfg",
				Usage: "Loads, expands the template arguments and prints the current config file",
				Action: func(ctx context.Context, c *cli.Command) error {
					if !config.HasConfig {
						fmt.Printf("The current direcory, does not contain a config file or it can not be loaded")
						return nil
					}

					b, err := toml.Marshal(config.CFG)
					if err != nil {
						return err
					}

					fmt.Print(string(b))

					return nil
				},
			},
			{
				Name:  "random-cfg",
				Usage: "Creates, marshals and prints a full config file with randomized data",
				Action: func(ctx context.Context, c *cli.Command) error {
					cfg := config.Cfg{}
					config.FillRandom(&cfg)
					b, _ := toml.Marshal(cfg)
					fmt.Print(string(b))

					return nil
				},
			},
		},
	},
	// {
	// 	Name:  "sync",
	// 	Usage: "Sync version numbers across config files and git tags",
	// 	Flags: []cli.Flag{
	// 		&cli.BoolFlag{
	// 			Name:  "revert",
	// 			Usage: "Revert sync if deployment fails",
	// 		},
	// 	},
	// 	Action: func(ctx context.Context, cmd *cli.Command) error {
	// 		revert := cmd.Bool("revert")
	// 		if revert {
	// 			fmt.Println("Reverting last sync...")
	// 			// TODO: rollback version bump + tags
	// 		} else {
	// 			fmt.Println("Syncing version across project...")
	// 			// TODO: update files, create git tag
	// 		}
	// 		return nil
	// 	},
	// },
	// {
	// 	Name:  "version",
	// 	Usage: "Manage project version",
	// 	Commands: []*cli.Command{
	// 		{
	// 			Name:  "show",
	// 			Usage: "Show current version",
	// 			Action: func(ctx context.Context, cmd *cli.Command) error {
	// 				fmt.Println("Current version: v0.1.0") // TODO: read from config
	// 				return nil
	// 			},
	// 		},
	// 		{
	// 			Name:  "bump",
	// 			Usage: "Bump version (major, minor, patch)",
	// 			Flags: []cli.Flag{
	// 				&cli.StringFlag{
	// 					Name:  "type",
	// 					Value: "patch",
	// 					Usage: "Type of bump (major, minor, patch)",
	// 				},
	// 			},
	// 			Action: func(ctx context.Context, cmd *cli.Command) error {
	// 				bumpType := cmd.String("type")
	// 				fmt.Printf("Bumping version (%s)...\n", bumpType)
	// 				// TODO: update config + tag
	// 				return nil
	// 			},
	// 		},
	// 	},
	// },
	// {
	// 	Name:  "release",
	// 	Usage: "Create a release (build + tag + publish)",
	// 	Action: func(ctx context.Context, cmd *cli.Command) error {
	// 		fmt.Println("Releasing project...")
	// 		// TODO: run build, bump version, push tags, publish artifacts
	// 		return nil
	// 	},
	// },
}

var err error

func main() {
	config.CFG_unexpanded, err = config.GetConfig()
	if err == nil {
		config.HasConfig = true
	}

	config.CFG, err = templating.ExpandConfig(config.CFG_unexpanded)
	if err != nil {
		log.Fatalf("could not expand templating arguments in config: %s", err)
	}

	if config.HasConfig {
		build_cmds = build.GenerateBuildCmds(config.CFG)
		if len(build_cmds) > 0 {
		}
	} else {
		build_action = func(ctx context.Context, cmd *cli.Command) error {
			return fmt.Errorf("'krill run' is not supproted without a config, use 'krill init' first")
		}
		build_cmds = nil
	}

	for _, c := range cmds {
		if c.Name == "run" {
			if build_cmds != nil {
				c.Commands = build_cmds
				c.DefaultCommand = build_cmds[0].Name
			} else {
				c.Action = build_action
			}
		}
	}

	cmd := &cli.Command{
		Name:     "krill",
		Usage:    "A simple language agnostic project manager, to make using other tools more pleasant",
		Commands: cmds,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "yes",
				Aliases: []string{"y"},
				Usage:   "run in non interactive mode, all yes/no prompts will assume a 'yes' anwser",
			},
			&cli.BoolFlag{
				Name:    "no",
				Aliases: []string{"n"},
				Usage:   "run in non interactive mode, all yes/no prompts will assume a 'no' anwser",
			},
		},
		Before: func(ctx context.Context, c *cli.Command) (context.Context, error) {
			cli_utils.SkipNO = c.Bool("no")
			cli_utils.SkipYES = c.Bool("yes")

			return ctx, nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
