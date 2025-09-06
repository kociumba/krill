# Commands

## `krill init`

Initialize a new project and create a `krill.toml` config. Detects language and build tool if possible.

---

## `krill run <target>`

Run the commands for a specific target (e.g. `debug`, `release`).  
Example:

```sh
krill run debug
```

Lists available targets if none are specified.

---

## `krill status`

Show project name, version, and config status. Also shows git status if available.

---

## `krill doctor [--auto-fix] [--diff]`

Check for issues in your config or environment.  
- `--auto-fix`: Apply suggested fixes automatically (**🚨 DESTRUCTIVE, use with caution**)).
- `--diff`: Show what would be changed in the suggested fix.

---

## `krill debug <command>`

Debugging utilities.  
Available subcommands:
- `expand-cfg`: Print the expanded config with all template values.
- `random-cfg`: Print a randomly generated config.

---

## Shell Completions

You can register command completions for: bash, zsh, fish and powershell. This is done by running `krill completion <shell>`.

Then sourcing the output in your shell to provide dynamic autocompletion results for krill, including dynamically suggesting availible target names.

More info on how to specifically source this output for each shell is availible on the [urfave/cli docs](https://cli.urfave.org/v3/examples/completions/shell-completions/)

---

For more details on each command, use `krill <command> --help`.
