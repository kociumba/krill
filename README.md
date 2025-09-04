# krill

<img align="right" src="https://raw.githubusercontent.com/kociumba/krill/main/assets/krill_icon.svg" alt="krill icon" width="150" height="150"/>

krill is a small project manager that helps you build, run, and manage projects across different languages and build systems. It works by reading a simple config file and providing a few commands to make common tasks easier.

## What does it do?

- Initializes a project config (`krill init`)
- Runs the commands defined under a specific target (`krill run <target>`)
- Shows project status (`krill status`)
- Checks your setup and config for issues, can automatically suggest fixes (`krill doctor`)
- All debug functionality, useful if anything works unexpectedly (`krill debug <command>`)

To get more info on any of these simply use `-h` or `--help`, which is contextual and works on each command separately.

## Why use it?

If you're tired of writing custom build scripts or remembering long build commands for every project, krill gives you a single config and a few commands to handle it all. It tries to detect your build system and language, and generates a starting config for you.

This auto generated config is more than enough for simple projects, but more mature ones will probably want to edit them with custom requirments.

## Quickstart

Install krill:

```bash
go install github.com/kociumba/krill
```

Initialize in your project (some functionality does not require a config):

```bash
krill init
```

Run a target (like the default `debug` or `release`):

```bash
krill run target-name
```

## Example: CMake Project

A generated `krill.toml` for a C/C++ project using CMake might look like:

```toml
[project]
name = "detected_project"
binary_type = "Executable"
languages = ["C", "Cpp"]
tools = ["CMake"]
version = "0.0.0"

[env.windows]
path = "powershell.exe"
args = ["-NoProfile", "-Command", "& { . 'C:\\...\\Launch-VsDevShell.ps1' -Arch amd64 }"]

[targets]
[targets.debug]
    output_dir = "cmake-build-debug"
    commands = [
      "cmake -S . -B {{ .targets.debug.output_dir }} -DCMAKE_BUILD_TYPE=Debug",
      "cmake --build {{ .targets.debug.output_dir }}"
    ]

[targets.release]
    output_dir = "cmake-build-release"
    commands = [
      "cmake -S . -B {{ .targets.release.output_dir }} -DCMAKE_BUILD_TYPE=Release",
      "cmake --build {{ .targets.release.output_dir }}"
    ]
```

`krill init` will try to find and generate the default enviornment for your platform, this means it will generate:
- powershell or cmd with vs developer setup, if you are on windows and the project uses C or C++
- powershell on windows with default arguments
- zsh on macOS with default arguments
- bash on linux with default arguments

A more mature, edited config for the same project might look something like this:

```toml
[project]
name = "detected_project"
binary_type = "Executable"
languages = ["C", "Cpp"]
tools = ["CMake"]
version = "1.6.5"

[env.windows]
path = "powershell.exe"
args = ["-NoProfile", "-Command", "& { . 'C:\\...\\Launch-VsDevShell.ps1' -Arch amd64 }"]

[targets]
[targets.run]
    depends_on = ["debug"]
    commands = ["{{ .targets.debug.output_dir }}/{{ .project.name }}{{ .exe_ext }}"]

[targets.debug]
    output_dir = "cmake-build-debug"
    commands = [
      "cmake -S . -B {{ .targets.debug.output_dir }} -DCMAKE_BUILD_TYPE=Debug",
      "cmake --build {{ .targets.debug.output_dir }}"
    ]

[targets.release]
    output_dir = "cmake-build-release"
    commands = [
      "cmake -S . -B {{ .targets.release.output_dir }} -DCMAKE_BUILD_TYPE=Release 
      -DBUILD_SHARED_LIBS=OFF",
      "cmake --build {{ .targets.release.output_dir }}"
    ]

[nested]
[nested.sub_project]
mappings = { run = "custom-run-target", debug = "custom-debug-target", release = "custom-release-target" }
```

> [!NOTE]
> Sub projects are supported by providing a `krill.toml` in the subdirectory, which then provides targets for the sub project.

## Supported

- Languages: C, C++, C#, F#, Go, Java, Kotlin, Odin, Rust
- Tools: CMake, Cargo, DotNet, Go, Gradle, Make, Meson, Nob, Taskfile, and direct compiler support (GCC, Clang, MSVC, JavaC, etc.)

If your tool or language is not supported by krill, you can use:

```toml
languages = ["CustomLang"]
tools = ["CustomTool"]
```

and since krill only automatically executes a commands for a target in the provided environment you can set up your build and targets however you want.

> [!WARNING]
> Support for raw compiler tools is quite unfinished right now, as they can not be supported the same way cmake or meson can be out of the box. But as said above anything that works with a command will also work with krill, using a custom target

---

For more info, run `krill --help` or look take a look at the [docs](https://kociumba.github.io/krill/).

to see all available language and tools names that krill recognizes, look in [./config/build_systems.go](./config/build_systems.go)

to see what the default generated builds for each tool are, take a look in [./build/generate_build.go](./build/generate_build.go) 
