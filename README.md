# krill

<img align="right" src="https://raw.githubusercontent.com/kociumba/krill/main/assets/krill_icon.svg" alt="krill icon" width="150" height="150"/>

A meta project manager to simplify your workflow.

krill sits on top of your existing build systems, providing a single, unified interface to build, run, and manage any project, regardless of the language or toolchain.

## The Problem

Let's be honest, using build systems like CMake directly can be a pain. Remembering, typing, and saving long, complex commands is tedious.

For example, configuring an average CMake project often looks something like this:

```bash
cmake -S . -B cmake-build-release \
  -DCMAKE_BUILD_TYPE=Release \
  -DCMAKE_INSTALL_PREFIX=/usr/local/my-app \
  -DENABLE_TESTS=ON \
  -DENABLE_EXAMPLES=OFF \
  -DAPPLICATION_VERSION=1.2.3 \
  -DSOME_LIBRARY_PATH=/usr/lib/some-library

cmake --build cmake-build-release
```

Many developers solve this by writing custom `build.sh` or `build.py` scripts, but that introduces new dependencies and boilerplate for every project.

## krills solution

krill replaces those messy scripts with a simple, declarative configuration file. It detects your project's tools and languages and generates a basic starting point, allowing you to manage your entire workflow with simple commands.

With krill, the annoying cmake command from above becomes just:

```bash
# Initialize your project with defaults (only needs to be done once)
krill init

# Build the release target
krill build release
```

### Core Features
- Auto-Detection: krill automatically detects your language and build system (CMake, Cargo, Gradle, etc.) to generate a working configuration.

- Simple, Declarative Config: Define your build targets in a straightforward TOML file. No more scripting required.

- Unified Interface: Use the same commands across all types of projects (C++, Rust, Java, Go, and more).

- Environment Management: builds requireing special enviornments are easely setup using project enviornments.

## How It Works

krill works by reading a `krill.toml` file in your project's root directory. When you run `krill init`, it scans for markers like `CMakeLists.txt`, `Cargo.toml`, or `go.mod` and creates a `krill.toml` file for you with basic defaults for the detected tools and languages.

Here's what a generated config for a C/C++ CMake project might look like:

```toml
# krill.toml

[project]
  name = "my-awesome-project"
  binary_type = "Executable"
  languages = ["C", "Cpp"]
  tools = ["CMake"]
  version = "0.0.0"

# Example of setting up the Visual Studio build environment on Windows
[env]
  path = "powershell.exe"
  args = ["-NoProfile", "-Command", "& { . 'C:\\...\\Launch-VsDevShell.ps1' -Arch amd64 }"]

# Define your build targets
[targets]
  [targets.debug]
    # Use templating to reference other values in the config
    output_dir = "build/debug"
    commands = [
      "cmake -S . -B {{ .targets.debug.output_dir }} -DCMAKE_BUILD_TYPE=Debug",
      "cmake --build {{ .targets.debug.output_dir }}"
    ]
    
  [targets.release]
    output_dir = "build/release"
    commands = [
      "cmake -S . -B {{ .targets.release.output_dir }} -DCMAKE_BUILD_TYPE=Release",
      "cmake --build {{ .targets.release.output_dir }}"
    ]
```

## Supported Platforms

krill aims to support a wide variety of languages and their common build tools out of the box.

- Languages: C, C++, C#, F#, Go, Java, Kotlin, Odin, Rust
- Tools: CMake, Cargo, DotNet, Go, Gradle, Make, Meson, Nob, Taskfile, and direct compiler support (GCC, Clang, MSVC, JavaC, etc.) for simple projects.
