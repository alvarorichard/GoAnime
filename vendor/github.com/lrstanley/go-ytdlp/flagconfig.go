// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in the
// LICENSE file.

package ytdlp

import (
	"errors"
	"reflect"
)

// UnmarshalRepair decodes data with [FlagConfig.UnmarshalJSONWithWarnings],
// then [FlagConfig.Repair]. Empty or whitespace input is treated as {}.
//
// On a hard parse failure of the root value, the receiver is not modified
// (same contract as [FlagConfig.UnmarshalJSONWithWarnings]).
func (f *FlagConfig) UnmarshalRepair(data []byte, opts ...FlagUnmarshalOption) ([]FlagWarning, error) {
	if f == nil {
		return nil, errors.New("cannot decode FlagConfig into a nil receiver")
	}

	warnings, err := f.UnmarshalJSONWithWarnings(data, opts...)
	if err != nil {
		return nil, err
	}

	return append(warnings, f.Repair()...), nil
}

// Repair unsets every field whose `id` tag matches a validation conflict
// reported by [FlagConfig.Validate], converting those errors to warnings.
// [FlagConfig.Validate] reports JSONPath as "<group>."+id (not the json
// struct tag); callers should repair with [FlagConfig.UnsetByID], not by
// deleting JSON map keys.
func (f *FlagConfig) Repair() []FlagWarning {
	if f == nil {
		return nil
	}

	err := f.Validate()
	if err == nil {
		return nil
	}

	flags := JSONParsingFlagErrors(err)
	if len(flags) == 0 {
		return []FlagWarning{{JSONPath: "$", Reason: err.Error()}}
	}

	var warnings []FlagWarning
	seen := make(map[string]struct{}, len(flags))
	for _, e := range flags {
		warnings = append(warnings, flagWarningFromParseErr(e))
		if e.ID == "" {
			continue
		}
		if _, ok := seen[e.ID]; ok {
			continue
		}
		seen[e.ID] = struct{}{}
		f.UnsetByID(e.ID)
	}
	return warnings
}

// UnsetByID zeros every field whose `id` tag matches id, including
// conflicting siblings that share one ID. Unknown IDs are ignored.
func (f *FlagConfig) UnsetByID(id string) {
	if f == nil || id == "" {
		return
	}
	unsetFieldsByID(reflect.ValueOf(f).Elem(), id)
}

// Overlay returns a copy of f with non-zero fields from over applied.
// Pointer fields are copied when non-nil; slices when non-empty. Nil
// overlay fields do not clear f. f and over are not modified.
func (f *FlagConfig) Overlay(over *FlagConfig) *FlagConfig {
	if f == nil {
		f = &FlagConfig{}
	}
	out := f.Clone()
	if over == nil {
		return out
	}
	overlayValue(reflect.ValueOf(out).Elem(), reflect.ValueOf(over.Clone()).Elem())
	return out
}

// OverlayFlagConfig repairs base and overlay, then [FlagConfig.Overlay].
func OverlayFlagConfig(base, overlay []byte) (*FlagConfig, []FlagWarning, error) {
	var b, o FlagConfig
	warnings, err := b.UnmarshalRepair(base)
	if err != nil {
		return nil, warnings, err
	}

	ow, err := o.UnmarshalRepair(overlay)
	warnings = append(warnings, ow...)
	if err != nil {
		return nil, warnings, err
	}

	merged := b.Overlay(&o)
	warnings = append(warnings, merged.Repair()...)
	return merged, warnings, nil
}

func flagWarningFromParseErr(e *ErrJSONParsingFlag) FlagWarning {
	reason := "invalid flag"
	if e.Err != nil {
		reason = e.Err.Error()
	}
	return FlagWarning{
		JSONPath: e.JSONPath,
		Flag:     e.Flag,
		ID:       e.ID,
		Reason:   reason,
	}
}

func unsetFieldsByID(v reflect.Value, id string) {
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return
	}

	t := v.Type()
	for i := range t.NumField() {
		field := t.Field(i)
		fv := v.Field(i)
		if !fv.CanSet() {
			continue
		}
		if field.Tag.Get("id") == id {
			fv.SetZero()
			continue
		}
		unsetFieldsByID(fv, id)
	}
}

func overlayValue(dst, src reflect.Value) {
	if !dst.IsValid() || !src.IsValid() || !dst.CanSet() {
		return
	}

	switch src.Kind() { //nolint:exhaustive // FlagConfig fields are pointers, slices, or structs
	case reflect.Pointer:
		if !src.IsNil() {
			dst.Set(src)
		}
	case reflect.Slice:
		if src.Len() > 0 {
			dst.Set(src)
		}
	case reflect.Struct:
		for i := range src.NumField() {
			overlayValue(dst.Field(i), src.Field(i))
		}
	default:
		if !src.IsZero() {
			dst.Set(src)
		}
	}
}
