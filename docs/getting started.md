# Getting Started

## Install

```bash
go install github.com/kociumba/krill@latest
```

## Initialize a Project

In your project directory:

```bash
krill init
```

This creates a `krill.toml` config file.

## Run a Target

To run a build or other target:

```bash
krill run <target>
```

For example, to build in debug mode:

```bash
krill run debug
```

## Next Steps

- Edit `krill.toml` to customize targets and settings.
- See [[config.md]] for details.
