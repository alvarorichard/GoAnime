// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

// Package optiondata contains the raw option data for go-ytdlp. Contents of this
// package are generated via cmd/codegen, and may change at any time.
package optiondata

import (
	"encoding/json"
	"fmt"
	"strings"
)

import _ "embed"

//go:embed json-schema.json
var JSONSchema []byte

// OpenAPIComponents returns the schemas from [JSONSchema] in the format
// expected by OpenAPI 3.1 components.schemas.
//
// JSON Schema definitions are moved from $defs to the returned map, and
// references to those definitions are rewritten from #/$defs/ to
// #/components/schemas/. The conversion is performed on parsed JSON values,
// rather than by replacing text, so unrelated strings containing that text
// are left unchanged.
func OpenAPIComponents() (map[string]json.RawMessage, error) {
	var document struct {
		Definitions map[string]json.RawMessage `json:"$defs"`
	}
	if err := json.Unmarshal(JSONSchema, &document); err != nil {
		return nil, fmt.Errorf("decode JSON schema: %w", err)
	}

	if document.Definitions == nil {
		return nil, fmt.Errorf("decode JSON schema: missing $defs")
	}

	components := make(map[string]json.RawMessage, len(document.Definitions))
	for name, raw := range document.Definitions {
		var value any
		if err := json.Unmarshal(raw, &value); err != nil {
			return nil, fmt.Errorf("decode JSON schema definition %q: %w", name, err)
		}

		rewriteOpenAPIRefs(value)
		converted, err := json.Marshal(value)
		if err != nil {
			return nil, fmt.Errorf("encode OpenAPI schema %q: %w", name, err)
		}
		components[name] = converted
	}

	return components, nil
}

func rewriteOpenAPIRefs(value any) {
	switch value := value.(type) {
	case map[string]any:
		if ref, ok := value["$ref"].(string); ok {
			if suffix, found := strings.CutPrefix(ref, "#/$defs/"); found {
				value["$ref"] = "#/components/schemas/" + suffix
			}
		}
		for _, child := range value {
			rewriteOpenAPIRefs(child)
		}
	case []any:
		for _, child := range value {
			rewriteOpenAPIRefs(child)
		}
	}
}

// OptionGroup is a group of options (e.g. general, verbosity, etc).
type OptionGroup struct {
	// Name of the option group.
	Name string `json:"name"`
	// Description of the option group, if any.
	Description string `json:"description,omitempty"`
	// Options are the options within the group.
	Options []*Option `json:"options"`
}

// Option is the raw option data for the given option (flag, essentially).
type Option struct {
	// ID is the identifier for the option, if one exists (may not for executables).
	// Note that this ID is not unique, as multiple options can have the same ID
	// (e.g. --something and --no-something).
	ID string `json:"id,omitempty"`
	// Name is the simplified name, based off the first found flags.
	Name string `json:"name"`
	// NameCamelCase is the same as [Option.Name], but in camelCase.
	NameCamelCase string `json:"name_camel_case"`
	// NamePascalCase is the same as [Option.Name], but in PascalCase.
	NamePascalCase string `json:"name_pascal_case"`
	// NameSnakeCase is the same as [Option.Name], but in snake_case.
	NameSnakeCase string `json:"name_snake_case"`
	// Links are optional links to the documentation for the option.
	URLs []*OptionURL `json:"urls,omitempty"`
	// DefaultFlag is the first flag (priority on long flags).
	DefaultFlag string `json:"default_flag"`
	// ArgNames are the argument names, if any -- length should match [Option.NArgs].
	ArgNames []string `json:"arg_names,omitempty"`
	// Executable is true if the option doesn't accept arguments.
	Executable bool `json:"executable"`
	// Deprecated will contain the deprecation description if the option if deprecated.
	Deprecated string `json:"deprecated,omitempty"`
	// Choices contains the list of required inputs for the option, if the option
	// has restricted inputs.
	Choices []string `json:"choices"`
	// Help contains the help text for the option.
	Help string `json:"help,omitempty"`
	// Hidden is true if the option is not returned in the help output (but can
	// still be provided).
	Hidden bool `json:"hidden"`
	// MetaArgs are the simplified syntax for the option, if any.
	MetaArgs string `json:"meta_args,omitempty"`
	// Type is the type (string, int, float64, bool, etc) of the option.
	Type string `json:"type"`
	// LongFlags are the extended flags for the option (e.g. --version).
	LongFlags []string `json:"long_flags"`
	// ShortFlags are the shortened flags for the option (e.g. -v).
	ShortFlags []string `json:"short_flags"`
	// NArgs is the number of arguments the option accepts.
	NArgs int `json:"nargs"`
	// NoOverride is true if the option should not override other flags with the same ID.
	NoOverride bool `json:"no_override"`
}

type OptionURL struct {
	// Name is the name of the option link.
	Name string `json:"name"`
	// URL is the link to the documentation for the option.
	URL string `json:"url"`
}
