package config

import (
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"slices"

	"github.com/BurntSushi/toml"
)

var cfg_file = "krill.toml"
var HasConfig = false
var CFG Cfg

type Cfg struct {
	Project      Project                  `toml:"project,omitempty"`
	Env          Environment              `toml:"env,omitempty"`
	BuildTargets map[string]BuildTarget   `toml:"targets,omitempty"`
	Nested       map[string]NestedProject `toml:"nested,omitempty"`
}

type Project struct {
	Name       string     `toml:"name,omitempty"`
	BinaryType BinaryType `toml:"binary_type,omitempty"`
	Languages  []Language `toml:"languages,omitempty"`
	Tools      []Tool     `toml:"tools,omitempty"`
	Version    string     `toml:"version,omitempty"`
}

type Environment struct {
	Path string   `toml:"path,omitempty"`
	Args []string `toml:"args,omitempty"`
}

type BuildTarget struct {
	Commands  []string `toml:"commands,omitempty"`
	OutputDir string   `toml:"output_dir,omitempty"`
	DependsOn []string `toml:"depends_on,omitempty"`
}

type NestedProject struct {
	Mappings map[string]string `toml:"mappings,omitempty"`
}

func EqualTools(a, b []Tool) bool {
	if len(a) != len(b) {
		return false
	}

	as, bs := slices.Clone(a), slices.Clone(b)
	slices.Sort(as)
	slices.Sort(bs)
	for i := range as {
		if as[i] != bs[i] {
			return false
		}
	}

	return true
}

func EqualLanguages(a, b []Language) bool {
	if len(a) != len(b) {
		return false
	}

	as, bs := slices.Clone(a), slices.Clone(b)
	slices.Sort(as)
	slices.Sort(bs)
	for i := range as {
		if as[i] != bs[i] {
			return false
		}
	}

	return true
}

func EqualEnv(a, b Environment) bool {
	if a.Path != b.Path {
		return false
	}

	if len(a.Args) != len(b.Args) {
		return false
	}

	for i := range a.Args {
		if a.Args[i] != b.Args[i] {
			return false
		}
	}

	return true
}

func EqualNested(a, b map[string]NestedProject) bool {
	if len(a) != len(b) {
		return false
	}

	for k, va := range a {
		vb, ok := b[k]
		if !ok {
			return false
		}

		if len(va.Mappings) != len(vb.Mappings) {
			return false
		}

		for mk, mv := range va.Mappings {
			if vb.Mappings[mk] != mv {
				return false
			}
		}
	}

	return true
}

func (p *Project) UnmarshalTOML(data any) error {
	m, ok := data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("expected map[string]interface{} for Project, got %T", data)
	}

	if name, ok := m["name"].(string); ok {
		p.Name = name
	}

	if version, ok := m["version"].(string); ok {
		p.Version = version
	}

	if langs, ok := m["languages"].([]interface{}); ok {
		p.Languages = make([]Language, len(langs))
		for i, lang := range langs {
			langStr, ok := lang.(string)
			if !ok {
				return fmt.Errorf("language at index %d is not a string: %T", i, lang)
			}

			var l Language
			if err := l.UnmarshalText([]byte(langStr)); err != nil {
				return fmt.Errorf("failed to unmarshal language %q: %w", langStr, err)
			}

			p.Languages[i] = l
		}
	}

	if tools, ok := m["tools"].([]interface{}); ok {
		p.Tools = make([]Tool, len(tools))
		for i, tool := range tools {
			toolStr, ok := tool.(string)
			if !ok {
				return fmt.Errorf("tool at index %d is not a string: %T", i, tool)
			}

			var t Tool
			if err := t.UnmarshalText([]byte(toolStr)); err != nil {
				return fmt.Errorf("failed to unmarshal tool %q: %w", toolStr, err)
			}

			p.Tools[i] = t
		}
	}

	return nil
}

func GetConfig() (Cfg, error) {
	if _, err := os.Stat(cfg_file); os.IsNotExist(err) || err != nil {
		return Cfg{}, fmt.Errorf("error opening or finding config file")
	}

	b, err := os.ReadFile(cfg_file)
	if err != nil {
		return Cfg{}, fmt.Errorf("error reading config file")
	}

	cfg := Cfg{}
	err = toml.Unmarshal(b, &cfg)
	if err != nil {
		return Cfg{}, fmt.Errorf("error unmarshaling config data")
	}

	return cfg, nil
}

func GetConfigFromDir(dir string) (Cfg, error) {
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(dir)
	return GetConfig()
}

func SaveConfig(cfg Cfg) error {
	if _, err := os.Stat(cfg_file); os.IsNotExist(err) {
		f, err := os.Create(cfg_file)
		if err != nil {
			return fmt.Errorf("failed to create config file: %w", err)
		}

		f.Close()
	}

	b, err := toml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	err = os.WriteFile(cfg_file, b, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func FillRandom(v any) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		panic("FillRandom requires a non-nil pointer")
	}

	fill(rv.Elem())
}

func fill(v reflect.Value) {
	switch v.Kind() {
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			f := v.Field(i)
			if !f.CanSet() {
				continue
			}

			fill(f)
		}
	case reflect.String:
		v.SetString(randString(8))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v.SetInt(rand.Int63n(100))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v.SetUint(uint64(rand.Intn(100)))
	case reflect.Slice:
		n := rand.Intn(3) + 1
		slice := reflect.MakeSlice(v.Type(), n, n)
		for i := range n {
			fill(slice.Index(i))
		}

		v.Set(slice)
	case reflect.Map:
		n := rand.Intn(3) + 1
		m := reflect.MakeMap(v.Type())
		for range n {
			key := reflect.New(v.Type().Key()).Elem()
			fill(key)
			val := reflect.New(v.Type().Elem()).Elem()
			fill(val)
			m.SetMapIndex(key, val)
		}

		v.Set(m)
	case reflect.Pointer:
		elem := reflect.New(v.Type().Elem())
		fill(elem.Elem())
		v.Set(elem)
	}
}

func randString(n int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
