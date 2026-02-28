package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// DetectProjectName tries to infer the project name from common project
// manifest files in dir. It checks pyproject.toml, package.json, and
// Cargo.toml in that order, returning the first non-empty name found.
// Falls back to the directory base name if no manifest provides a name.
// Errors from manifest files are silently ignored.
func DetectProjectName(dir string) string {
	if name := detectFromPyproject(dir); name != "" {
		return name
	}
	if name := detectFromPackageJSON(dir); name != "" {
		return name
	}
	if name := detectFromCargo(dir); name != "" {
		return name
	}
	return filepath.Base(dir)
}

type pyprojectTOML struct {
	Project struct {
		Name string `toml:"name"`
	} `toml:"project"`
	Tool struct {
		Poetry struct {
			Name string `toml:"name"`
		} `toml:"poetry"`
	} `toml:"tool"`
}

func detectFromPyproject(dir string) string {
	var p pyprojectTOML
	if _, err := toml.DecodeFile(filepath.Join(dir, "pyproject.toml"), &p); err != nil {
		return ""
	}
	if p.Project.Name != "" {
		return p.Project.Name
	}
	return p.Tool.Poetry.Name
}

type packageJSON struct {
	Name string `json:"name"`
}

func detectFromPackageJSON(dir string) string {
	data, err := os.ReadFile(filepath.Join(dir, "package.json"))
	if err != nil {
		return ""
	}
	var p packageJSON
	if err := json.Unmarshal(data, &p); err != nil {
		return ""
	}
	return p.Name
}

type cargoTOML struct {
	Package struct {
		Name string `toml:"name"`
	} `toml:"package"`
}

func detectFromCargo(dir string) string {
	var c cargoTOML
	if _, err := toml.DecodeFile(filepath.Join(dir, "Cargo.toml"), &c); err != nil {
		return ""
	}
	return c.Package.Name
}
