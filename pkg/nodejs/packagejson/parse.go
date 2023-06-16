package packagejson

import (
	"encoding/json"
	"io"

	"github.com/aquasecurity/go-dep-parser/pkg/types"
	"github.com/aquasecurity/go-dep-parser/pkg/utils"
	"golang.org/x/xerrors"
)

type packageJSON struct {
	Name                 string            `json:"name"`
	Version              string            `json:"version"`
	License              interface{}       `json:"license"`
	Dependencies         map[string]string `json:"dependencies"`
	OptionalDependencies map[string]string `json:"optionalDependencies"`
}

func (p packageJSON) hasContent() bool {
	return parseLicense(p.License) != "" || len(p.Dependencies) > 0 || len(p.OptionalDependencies) > 0
}

type Package struct {
	types.Library
	Dependencies         map[string]string
	OptionalDependencies map[string]string
}

type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) Parse(r io.Reader) (Package, error) {
	var pkgJSON packageJSON
	if err := json.NewDecoder(r).Decode(&pkgJSON); err != nil {
		return Package{}, xerrors.Errorf("JSON decode error: %w", err)
	}

	if pkgJSON.Name == "" && pkgJSON.Version == "" && !pkgJSON.hasContent() {
		return Package{}, nil
	}

	name := pkgJSON.Name
	// Name and version fields are optional
	// https://docs.npmjs.com/cli/v9/configuring-npm/package-json#name
	// In cases where the name is not provided, but there is valuable content present,
	// we utilize a placeholder constant to identify this package.
	if name == "" {
		name = "no-package-name"
	}

	return Package{
		Library: types.Library{
			ID:      utils.PackageID(name, pkgJSON.Version),
			Name:    name,
			Version: pkgJSON.Version,
			License: parseLicense(pkgJSON.License),
		},
		Dependencies:         pkgJSON.Dependencies,
		OptionalDependencies: pkgJSON.OptionalDependencies,
	}, nil
}

func parseLicense(val interface{}) string {
	// the license isn't always a string, check for legacy struct if not string
	switch v := val.(type) {
	case string:
		return v
	case map[string]interface{}:
		if license, ok := v["type"]; ok {
			return license.(string)
		}
	}
	return ""
}
