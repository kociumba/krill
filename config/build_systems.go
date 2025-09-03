package config

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"runtime"
	"strings"
)

//go:generate stringer -type=Language
type Language int

//go:generate stringer -type=Tool
type Tool int

//go:generate stringer -type=BinaryType
type BinaryType int

const (
	C Language = iota
	Cpp
	Kotlin
	Java
	Rust
	Go
	Odin
	CSharp
	FSharp
	CustomLang
)

func (l Language) MarshalText() ([]byte, error) {
	return []byte(l.String()), nil
}

func (l *Language) UnmarshalText(text []byte) error {
	for i := range Language(len(_Language_index) - 1) {
		if i.String() == string(text) {
			*l = i
			return nil
		}
	}

	return fmt.Errorf("unknown language: %s", text)
}

const (
	CMake Tool = iota
	Nob
	Raw_GCC  // assumed when on linux and no other build system is detectd in a c/c++ project
	Raw_MSVC // assumed when on windows and no other build system is detectd in a c/c++ project
	raw_CLANG
	Gradle
	raw_KotlinC // assumed when no gradle is detected in a kotlin project
	raw_JavaC   // assumed when no gradle is detected in a java project
	Meson
	Cargo
	raw_RustC // assumed when no cargo is detected in a rust project
	Make
	Taskfile
	GoCmd
	OdinCmd
	DotNet
	CustomTool
)

func (t Tool) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

func (t *Tool) UnmarshalText(text []byte) error {
	for i := range Tool(len(_Tool_index) - 1) {
		if i.String() == string(text) {
			*t = i
			return nil
		}
	}

	return fmt.Errorf("unknown tool: %s", text)
}

const (
	Executable BinaryType = iota
	SharedLib
	DynamicLib
	StaticLib
	Object
	Framework
)

func (b BinaryType) MarshalText() ([]byte, error) {
	return []byte(b.String()), nil
}

func (b *BinaryType) UnmarshalText(text []byte) error {
	for i := range BinaryType(len(_BinaryType_index) - 1) {
		if i.String() == string(text) {
			*b = i
			return nil
		}
	}

	return fmt.Errorf("unknown tool: %s", text)
}

var BinaryTypeToExt map[BinaryType]string

func populateBinaryTypeToExt() {
	BinaryTypeToExt = make(map[BinaryType]string)
	switch runtime.GOOS {
	case "windows":
		BinaryTypeToExt[Executable] = ".exe"
		BinaryTypeToExt[DynamicLib] = ".dll"
		BinaryTypeToExt[StaticLib] = ".lib"
		BinaryTypeToExt[Object] = ".obj"
		BinaryTypeToExt[SharedLib] = ".dll"
	case "darwin":
		BinaryTypeToExt[Executable] = ""
		BinaryTypeToExt[DynamicLib] = ".dylib"
		BinaryTypeToExt[StaticLib] = ".a"
		BinaryTypeToExt[Object] = ".o"
		BinaryTypeToExt[SharedLib] = ".dylib"
		BinaryTypeToExt[Framework] = ".framework"
	default: // Linux and other Unix-like systems
		BinaryTypeToExt[Executable] = ""
		BinaryTypeToExt[DynamicLib] = ".so"
		BinaryTypeToExt[StaticLib] = ".a"
		BinaryTypeToExt[Object] = ".o"
		BinaryTypeToExt[SharedLib] = ".so"
	}
}

var LangToTool = map[Language]map[Tool]struct{}{
	C:      {CMake: {}, Meson: {}, Make: {}, Taskfile: {}, Nob: {}, Raw_GCC: {}, Raw_MSVC: {}, raw_CLANG: {}},
	Cpp:    {CMake: {}, Meson: {}, Make: {}, Taskfile: {}, Raw_GCC: {}, Raw_MSVC: {}, raw_CLANG: {}},
	Kotlin: {Gradle: {}, Make: {}, Taskfile: {}, raw_KotlinC: {}},
	Java:   {Gradle: {}, Make: {}, Taskfile: {}, raw_JavaC: {}},
	Rust:   {Cargo: {}, Make: {}, Taskfile: {}, raw_RustC: {}},
	Go:     {GoCmd: {}, Taskfile: {}, Make: {}},
	Odin:   {OdinCmd: {}, Taskfile: {}, Make: {}},
	CSharp: {DotNet: {}, Taskfile: {}, Make: {}},
	FSharp: {DotNet: {}, Taskfile: {}, Make: {}},
}

var ToolToLang = map[Tool]map[Language]struct{}{}

func init() {
	for lang, tools := range LangToTool {
		for tool := range tools {
			if ToolToLang[tool] == nil {
				ToolToLang[tool] = map[Language]struct{}{}
			}
			ToolToLang[tool][lang] = struct{}{}
		}
	}

	populateBinaryTypeToExt()
}

var ToolMarkers = map[Tool][]string{
	CMake:    {"CMakeLists.txt", "*.cmake"},
	Nob:      {"nob", "nob.exe", "nob.h"},
	Gradle:   {"*.gradle", "*.gradle.kts"},
	Meson:    {"meson.build"},
	Cargo:    {"Cargo.toml"},
	Make:     {"Makefile", "makefile"},
	Taskfile: {"Taskfile.yml", "Taskfile.yaml", "taskfile.yml", "taskfile.yaml"},
	GoCmd:    {"go.mod", "go.sum", "*.go"},
	OdinCmd:  {"*.odin"},
	DotNet:   {"*.sln", "*.csproj", "*.fsproj", "*.fs", "*.cs"},
}

func DetectTools(root string) []Tool {
	entries, err := filepath.Glob(filepath.Join(root, "*"))
	if err != nil {
		return nil
	}

	files := map[string]struct{}{}
	for _, e := range entries {
		base := filepath.Base(e)
		files[base] = struct{}{}
	}

	var found []Tool
	for tool, patterns := range ToolMarkers {
		for _, pat := range patterns {
			if strings.Contains(pat, "*") {
				for f := range files {
					ok, _ := filepath.Match(pat, f)
					if ok {
						found = append(found, tool)
						break
					}
				}
			} else {
				if _, ok := files[pat]; ok {
					found = append(found, tool)
					break
				}
			}
		}
	}

	if len(found) == 0 {
		found = append(found, detectRawByExtension(root)...)
	}

	return found
}

func detectRawByExtension(root string) []Tool {
	var hasC, hasCpp, hasKotlin, hasJava, hasRust bool

	filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		switch filepath.Ext(d.Name()) {
		case ".c":
			hasC = true
		case ".cpp":
			hasCpp = true
		case ".kt":
			hasKotlin = true
		case ".java":
			hasJava = true
		case ".rs":
			hasRust = true
		}
		return nil
	})

	var out []Tool
	if hasC || hasCpp {
		var rawTool Tool
		switch runtime.GOOS {
		case "windows":
			rawTool = Raw_MSVC
		case "linux":
			rawTool = Raw_GCC
		case "darwin":
			rawTool = raw_CLANG
		default:
			rawTool = Raw_GCC
		}
		out = append(out, rawTool)
	}
	if hasKotlin {
		out = append(out, raw_KotlinC)
	}
	if hasJava {
		out = append(out, raw_JavaC)
	}
	if hasRust {
		out = append(out, raw_RustC)
	}
	return out
}

func DetectLanguages(root string, tools []Tool) []Language {
	if tools == nil {
		tools = DetectTools(root)
	}

	langs := map[Language]struct{}{}
	for _, t := range tools {
		for l := range ToolToLang[t] {
			langs[l] = struct{}{}
		}
	}

	out := make([]Language, 0, len(langs))
	for l := range langs {
		out = append(out, l)
	}

	return out
}
