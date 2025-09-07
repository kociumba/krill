# Configuration Reference

krill uses a `krill.toml` file in your project root. This file defines project info, build targets, environment, and nested projects.

---

## Example `krill.toml`

```toml
[project]
name = "my_project"
binary_type = "Executable"
languages = ["C", "Cpp"]
tools = ["CMake"]
version = "0.1.0"

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

[nested]
[nested.sub_project]
mappings = { debug = "custom-debug", release = "custom-release" }
```

---

## Sections

- `[project]`: Name, version, binary type, languages, tools.
- `[env]`: Command and arguments used to run build commands.
- `[targets]`: Build targets. Each target can have `commands`, `output_dir`, and `depends_on`.
- `[nested]`: Subprojects with their own `krill.toml`.

---

## Templating

Templating is supported, throught standard go tmpl syntax: `{{ .var }}`, the config file goes throught a one pass template expansion so nested and recursive templates are not supported, in addition to each variable defined in the config, special utility variables:

- `exe_ext` - provides the executable file extension on the executing platform
- `dll_ext` - provides the dynamic/shared library extension
- `static_lib_ext` - provides the static library extension
- `obj_ext` - provides the object file extension
- `shared_lib_ext` - provides the shared library extension (typically same as `dll_ext`)
- `framework_ext` - provides the framework extension if supported on the platform

---

## Supported Out of the Box

- Languages: C, C++, C#, F#, Go, Java, Kotlin, Odin, Rust
- Tools: CMake, Cargo, DotNet, Go, Gradle, Make, Meson, Nob, Taskfile
- Default targets: `debug` and `release` (auto-generated for known tools/languages)
- Nested projects (subdirectories with their own config)

---

## Not Supported

- Arbitrary scripting or hooks (only command lists per target)
- Automatic dependency management between unrelated projects
- Advanced build graph features (custom rules, file watching, etc.)

---

## Defaults

- If you run `krill init`, krill tries to detect your language and tool, and generates a config with default `debug` and `release` targets.
- Each supported tool has its own default commands for these targets, you can see them in [generate_build.go](https://github.com/kociumba/krill/blob/main/build/generate_build.go).
- If multiple tools are detected, targets are named like `debug-cmake`, `debug-cargo`, etc., and aggregate targets are created.
- The environment is set to a shell suitable for your platform (e.g., `powershell.exe` on Windows). This will try to find vs developer powershell or cmd on windows if the project contains C or C++.

> [!TIP]
> If you do not know what fields are supported in the config and what can be customized, you can always use `krill debug random-cfg`, this will print out a fully filled out config with random values, you can use that to see the overall shape of the structures and what is supported where.

---

For more details on targets and commands, see [[commands.md]].
