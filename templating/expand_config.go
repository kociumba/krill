package templating

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/BurntSushi/toml"
	"github.com/kociumba/krill/config"
)

func ExpandConfig(cfg config.Cfg) (config.Cfg, error) {
	wd, err := os.Getwd()
	if err != nil {
		return config.Cfg{}, fmt.Errorf("failed to get working directory: %w", err)
	}

	filePath := filepath.Join(wd, "krill.toml")
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return config.Cfg{}, fmt.Errorf("failed to read krill.toml: %w", err)
	}

	if !strings.Contains(string(fileContent), "{{") {
		return cfg, nil
	}

	templateData, err := resolveTags(cfg)
	if err != nil {
		return config.Cfg{}, fmt.Errorf("failed to resolve template tags: %w", err)
	}

	templateData["exe_ext"] = config.BinaryTypeToExt[config.Executable]
	templateData["dll_ext"] = config.BinaryTypeToExt[config.DynamicLib]
	templateData["static_lib_ext"] = config.BinaryTypeToExt[config.StaticLib]
	templateData["obj_ext"] = config.BinaryTypeToExt[config.Object]
	templateData["shared_lib_ext"] = config.BinaryTypeToExt[config.SharedLib]
	if config.BinaryTypeToExt[config.Framework] != "" {
		templateData["framework_ext"] = config.BinaryTypeToExt[config.Framework]
	}

	tmpl, err := template.New("config").Parse(string(fileContent))
	if err != nil {
		return config.Cfg{}, fmt.Errorf("failed to parse template: %w", err)
	}

	var sb strings.Builder
	err = tmpl.Execute(&sb, templateData)
	if err != nil {
		return config.Cfg{}, fmt.Errorf("failed to execute template: %w", err)
	}

	var newCfg config.Cfg
	err = toml.Unmarshal([]byte(sb.String()), &newCfg)
	if err != nil {
		return config.Cfg{}, fmt.Errorf("failed to unmarshal rendered TOML: %w", err)
	}

	return newCfg, nil
}
