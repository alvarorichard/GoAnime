// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package ytdlp

import (
	"bytes"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"maps"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"github.com/lrstanley/go-ytdlp/optiondata"
)

// FlagWarning describes a member that was unknown or could not be
// decoded while using [FlagConfig.UnmarshalJSONWithWarnings].
type FlagWarning struct {
	JSONPath string `json:"json_path"`
	Flag     string `json:"flag,omitempty"`
	ID       string `json:"id,omitempty"`
	Reason   string `json:"reason"`
}

// FlagCoercionFunc can provide a value for a [FlagConfig] member that
// cannot be decoded by the default compatibility rules. The returned value
// must be assignable to, or convertible to, target. Returning false leaves the
// value to the default decoder. The raw JSON must not be retained.
type FlagCoercionFunc func(path string, raw []byte, target reflect.Type) (value any, handled bool, err error)

// FlagUnmarshalOption configures compatibility decoding.
//
// The decoder type is private so callers can only use package-provided
// With* functions to create options.
type FlagUnmarshalOption func(dec *flagConfigDecoder) *flagConfigDecoder

// WithFlagCoercion returns options that use coerce to provide
// caller-defined compatibility conversions.
func WithFlagCoercion(coerce FlagCoercionFunc) FlagUnmarshalOption {
	return func(dec *flagConfigDecoder) *flagConfigDecoder {
		dec.coerce = coerce
		return dec
	}
}

type flagConfigDecoder struct {
	coerce   FlagCoercionFunc
	warnings []FlagWarning
}

type flagConfigFieldInfo struct {
	flag string
	id   string
}

// UnmarshalJSONWithWarnings decodes a [FlagConfig] while retaining valid
// members and returning warnings for unknown or incompatible members.
//
// Nil, empty, and whitespace-only input are treated as {}. A non-object
// root (array, null, or malformed JSON) is a hard error; the receiver is
// not modified.
//
// The receiver is updated in place. Members absent from data, and members
// that cannot be decoded, retain their existing values. Unknown members are
// reported but are not retained.
func (f *FlagConfig) UnmarshalJSONWithWarnings(data []byte, options ...FlagUnmarshalOption) ([]FlagWarning, error) {
	if f == nil {
		return nil, errors.New("cannot decode FlagConfig into a nil receiver")
	}

	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		data = []byte("{}")
	}

	if jsontext.Value(data).Kind() != jsontext.KindBeginObject {
		return nil, errors.New("flag config JSON root must be an object")
	}

	var members map[string]jsontext.Value
	if err := json.Unmarshal(data, &members); err != nil {
		return nil, fmt.Errorf("decode flag config JSON: %w", err)
	}

	decoder := &flagConfigDecoder{}
	for _, option := range options {
		if option != nil {
			decoder = option(decoder)
		}
	}
	if decoder == nil {
		return nil, errors.New("flag config decoder option returned a nil decoder")
	}
	decoder.decodeStruct(members, reflect.ValueOf(f).Elem(), "", flagConfigFieldInfo{})
	return decoder.warnings, nil
}

func (d *flagConfigDecoder) decodeStruct(
	members map[string]jsontext.Value,
	dst reflect.Value,
	path string,
	inherited flagConfigFieldInfo,
) {
	fields := make(map[string]reflect.StructField, dst.NumField())
	for i := range dst.NumField() {
		field := dst.Type().Field(i)
		name, _, _ := strings.Cut(field.Tag.Get("json"), ",")
		if name == "" {
			name = field.Name
		}
		if name != "-" {
			fields[name] = field
		}
	}

	for _, name := range slices.Sorted(maps.Keys(members)) {
		raw := members[name]
		field, ok := fields[name]
		if !ok {
			d.addWarning(joinFlagConfigPath(path, name), flagConfigMemberInfo(name, nil, inherited), "unknown member")
			continue
		}

		fieldInfo := flagConfigMemberInfo(name, &field, inherited)
		fieldPath := joinFlagConfigPath(path, name)
		fieldValue := dst.FieldByIndex(field.Index)
		if err := d.decodeValue(raw, fieldValue, fieldPath, fieldInfo); err != nil {
			d.addWarning(fieldPath, fieldInfo, err.Error())
		}
	}
}

func (d *flagConfigDecoder) decodeValue(
	raw jsontext.Value,
	dst reflect.Value,
	path string,
	info flagConfigFieldInfo,
) error {
	if d.coerce != nil {
		value, handled, err := d.coerce(path, raw, dst.Type())
		if err != nil {
			return fmt.Errorf("custom coercion: %w", err)
		}
		if handled {
			if assignErr := assignFlagConfigValue(dst, value); assignErr != nil {
				return fmt.Errorf("custom coercion: %w", assignErr)
			}
			return nil
		}
	}

	switch dst.Kind() { //nolint:exhaustive // unsupported kinds use the scalar decoder
	case reflect.Pointer:
		if raw.Kind() == jsontext.KindNull {
			dst.SetZero()
			return nil
		}

		value := reflect.New(dst.Type().Elem())
		if err := d.decodeValue(raw, value.Elem(), path, info); err != nil {
			return err
		}
		dst.Set(value)
		return nil
	case reflect.Struct:
		if raw.Kind() != jsontext.KindBeginObject {
			return d.decodeDirect(raw, dst)
		}

		var members map[string]jsontext.Value
		if err := json.Unmarshal(raw, &members); err != nil {
			return fmt.Errorf("decode object: %w", err)
		}
		d.decodeStruct(members, dst, path, info)
		return nil
	case reflect.Slice:
		return d.decodeSlice(raw, dst, path, info)
	default:
		return d.decodeScalar(raw, dst)
	}
}

func (d *flagConfigDecoder) decodeSlice(
	raw jsontext.Value,
	dst reflect.Value,
	path string,
	info flagConfigFieldInfo,
) error {
	if raw.Kind() == jsontext.KindNull {
		dst.SetZero()
		return nil
	}

	if raw.Kind() != jsontext.KindBeginArray {
		if !isPrimitiveJSONKind(raw.Kind()) {
			return fmt.Errorf("expected an array, got %s", raw.Kind())
		}

		value := reflect.New(dst.Type().Elem()).Elem()
		if err := d.decodeValue(raw, value, path+"[0]", info); err != nil {
			return err
		}
		result := reflect.MakeSlice(dst.Type(), 1, 1)
		result.Index(0).Set(value)
		dst.Set(result)
		return nil
	}

	var values []jsontext.Value
	if err := json.Unmarshal(raw, &values); err != nil {
		return fmt.Errorf("decode array: %w", err)
	}

	result := reflect.MakeSlice(dst.Type(), 0, len(values))
	var failed bool
	for index, value := range values {
		element := reflect.New(dst.Type().Elem()).Elem()
		elementPath := path + "[" + strconv.Itoa(index) + "]"
		if err := d.decodeValue(value, element, elementPath, info); err != nil {
			d.addWarning(elementPath, info, err.Error())
			failed = true
			continue
		}
		result = reflect.Append(result, element)
	}
	if !failed {
		dst.Set(result)
	}
	return nil
}

func (d *flagConfigDecoder) decodeScalar(raw jsontext.Value, dst reflect.Value) error {
	directErr := d.decodeDirect(raw, dst)
	if directErr == nil {
		return nil
	}

	value, ok, err := coerceFlagConfigScalar(raw, dst.Type())
	if !ok {
		return directErr
	}
	if err != nil {
		return err
	}
	dst.Set(value)
	return nil
}

func (d *flagConfigDecoder) decodeDirect(raw jsontext.Value, dst reflect.Value) error {
	if !dst.CanAddr() {
		return errors.New("target is not addressable")
	}
	if err := json.Unmarshal(raw, dst.Addr().Interface()); err != nil {
		return fmt.Errorf("cannot decode %s as %s: %w", raw.Kind(), dst.Type(), err)
	}
	return nil
}

func (d *flagConfigDecoder) addWarning(path string, info flagConfigFieldInfo, reason string) {
	d.warnings = append(d.warnings, FlagWarning{
		JSONPath: path,
		Flag:     info.flag,
		ID:       info.id,
		Reason:   reason,
	})
}

func flagConfigMemberInfo(name string, field *reflect.StructField, inherited flagConfigFieldInfo) flagConfigFieldInfo {
	info := inherited
	if field != nil {
		if id := field.Tag.Get("id"); id != "" {
			info.id = id
		}
	}

	if option := optiondata.FindByName(name); option != nil {
		if info.flag == "" {
			info.flag = option.DefaultFlag
		}
		if info.id == "" {
			info.id = option.ID
		}
	}
	return info
}

func joinFlagConfigPath(parent, name string) string {
	if parent == "" {
		return name
	}
	return parent + "." + name
}

func isPrimitiveJSONKind(kind jsontext.Kind) bool {
	return kind == jsontext.KindNull ||
		kind == jsontext.KindFalse ||
		kind == jsontext.KindTrue ||
		kind == jsontext.KindString ||
		kind == jsontext.KindNumber
}

func coerceFlagConfigScalar(raw jsontext.Value, target reflect.Type) (reflect.Value, bool, error) {
	value := strings.TrimSpace(string(raw))
	result := reflect.New(target).Elem()

	switch target.Kind() { //nolint:exhaustive // only JSON scalar kinds have compatibility conversions
	case reflect.String:
		switch raw.Kind() { //nolint:exhaustive // unsupported kinds are not string scalars
		case jsontext.KindNumber, jsontext.KindFalse, jsontext.KindTrue:
			result.SetString(value)
			return result, true, nil
		}
	case reflect.Bool:
		if raw.Kind() == jsontext.KindString {
			parsed, err := strconv.ParseBool(value[1 : len(value)-1])
			if err != nil {
				return reflect.Value{}, true, fmt.Errorf("cannot coerce %q to bool: %w", value, err)
			}
			result.SetBool(parsed)
			return result, true, nil
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if raw.Kind() == jsontext.KindString {
			parsed, err := strconv.ParseInt(value[1:len(value)-1], 10, target.Bits())
			if err != nil {
				return reflect.Value{}, true, fmt.Errorf("cannot coerce %q to %s: %w", value, target, err)
			}
			result.SetInt(parsed)
			return result, true, nil
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		if raw.Kind() == jsontext.KindString {
			parsed, err := strconv.ParseUint(value[1:len(value)-1], 10, target.Bits())
			if err != nil {
				return reflect.Value{}, true, fmt.Errorf("cannot coerce %q to %s: %w", value, target, err)
			}
			result.SetUint(parsed)
			return result, true, nil
		}
	case reflect.Float32, reflect.Float64:
		if raw.Kind() == jsontext.KindString {
			parsed, err := strconv.ParseFloat(value[1:len(value)-1], target.Bits())
			if err != nil {
				return reflect.Value{}, true, fmt.Errorf("cannot coerce %q to %s: %w", value, target, err)
			}
			result.SetFloat(parsed)
			return result, true, nil
		}
	}

	return reflect.Value{}, false, nil
}

func assignFlagConfigValue(dst reflect.Value, value any) error {
	if value == nil {
		switch dst.Kind() { //nolint:exhaustive // only nullable kinds accept nil
		case reflect.Pointer, reflect.Interface, reflect.Map, reflect.Slice:
			dst.SetZero()
			return nil
		default:
			return errors.New("custom coercion returned nil for a non-nullable target")
		}
	}

	source := reflect.ValueOf(value)
	if source.Type().AssignableTo(dst.Type()) {
		dst.Set(source)
		return nil
	}
	if source.Type().ConvertibleTo(dst.Type()) {
		dst.Set(source.Convert(dst.Type()))
		return nil
	}
	if dst.Kind() == reflect.Pointer && source.Type().AssignableTo(dst.Type().Elem()) {
		result := reflect.New(dst.Type().Elem())
		result.Elem().Set(source)
		dst.Set(result)
		return nil
	}
	return fmt.Errorf("value of type %s cannot be assigned to %s", source.Type(), dst.Type())
}
