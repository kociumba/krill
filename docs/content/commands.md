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

For more details on each command, use `krill <command> --help`.
