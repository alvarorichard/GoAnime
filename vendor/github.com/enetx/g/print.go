package g

import (
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
)

// Formattable lets a type handle g.Format specifications without reflection.
// The spec is the text after ':' without the colon; an empty spec requests the
// type's default representation.
type Formattable interface {
	FormatValue(spec String) String
}

// Write formats according to a format specifier and writes to w.
// It returns a Result containing the number of bytes written or an error.
//
// Example:
//
//	res := g.Write(os.Stdout, "Hello, {}!\n", "world")
//	if res.IsErr() { log.Fatal(res.Err()) }
func Write[T ~string](w io.Writer, format T, args ...any) Result[int] {
	return ResultOf(io.WriteString(w, Format(format, args...).Std()))
}

// Writeln formats according to a format specifier, appends a newline, and writes to w.
// It returns a Result containing the number of bytes written or an error.
//
// Example:
//
//	res := g.Writeln(os.Stdout, "Hello, {}", "world")
//	if res.IsErr() { log.Fatal(res.Err()) }
func Writeln[T ~string](w io.Writer, format T, args ...any) Result[int] {
	return ResultOf(io.WriteString(w, formatTemplate(format, args, "\n").Std()))
}

// Print formats according to a format specifier and writes to os.Stdout.
// It returns a Result containing the number of bytes written or an error.
//
// Example:
//
//	g.Print("Hello, {}!\n", "world")
func Print[T ~string](format T, args ...any) Result[int] {
	return Write(os.Stdout, format, args...)
}

// Println formats according to a format specifier, appends a newline, and writes to os.Stdout.
// It returns a Result containing the number of bytes written or an error.
//
// Example:
//
//	g.Println("Hello, {}", "world")
func Println[T ~string](format T, args ...any) Result[int] {
	return Writeln(os.Stdout, format, args...)
}

// Eprint formats according to a format specifier and writes to os.Stderr.
// It returns a Result containing the number of bytes written or an error.
//
// Example:
//
//	g.Eprint("Error: {}", "file not found")
func Eprint[T ~string](format T, args ...any) Result[int] {
	return Write(os.Stderr, format, args...)
}

// Eprintln formats according to a format specifier, appends a newline, and writes to os.Stderr.
// It returns a Result containing the number of bytes written or an error.
//
// Example:
//
//	g.Eprintln("Error: {}", "permission denied")
func Eprintln[T ~string](format T, args ...any) Result[int] {
	return Writeln(os.Stderr, format, args...)
}

// Errorf formats according to a format specifier and returns it as an error.
// If any argument is referenced via the {:w} format verb, it is both displayed
// and wrapped into the returned error, making errors.Is and errors.As work
// through the chain. Multiple {:w} references wrap multiple errors (Go 1.20+).
//
// Example:
//
//	err := g.Errorf("could not open {1}: {2:w}", filename, err)
//	errors.Is(err, os.ErrNotExist) // true
func Errorf[T ~string](format T, args ...any) error {
	tmpl := String(format)
	named, positional := formatArgs(args)

	var wraps []error

	msg := parseTmpl(tmpl, named, positional, &wraps)

	if len(wraps) == 0 {
		return errors.New(msg.Std())
	}

	return &wrappedError{msg: msg.Std(), errs: wraps}
}

// Format processes a template string and replaces placeholders with corresponding values from the provided arguments.
// It supports numeric, named, and auto-indexed placeholders, with dot-access into map keys, slice/array indices, and struct fields.
//
// If a placeholder cannot resolve a value, it remains unchanged in the output.
//
// Parameters:
//   - template (T ~string): A string containing placeholders enclosed in `{}`.
//   - args (...any): A variadic list of arguments, which may include:
//   - Positional arguments (numbers, strings, slices, structs, maps, etc.).
//   - A `Named` map for named placeholders.
//
// Placeholder Forms:
//   - Numeric: `{1}`, `{2}` - References positional arguments by their 1-based index.
//   - Named: `{key}`, `{key.field}`, `{key.0}` - References keys from a `Named` map with data access into struct fields, map/MapOrd keys, and slice indices (methods are never invoked).
//   - Fallback: `{key?fallback}` - If `key` is not found in the named map, uses the value
//     of the named key `fallback` instead (not the literal text).
//   - Auto-index: `{}` - Automatically uses the next positional argument if the placeholder is empty.
//   - Escaping: `{{` and `}}` - Emits literal braces.
//
// Returns:
//   - String: A formatted string with all resolved placeholders replaced by their corresponding values.
//
// Notes:
//   - If a placeholder cannot resolve a value (e.g., missing key or out-of-range index), it remains unchanged in the output.
//   - Dot-segments access data only (struct fields, map keys, MapOrd keys, slice/array indices); methods are never invoked. An unresolvable segment leaves the value unmodified.
//   - Only a single `Named` map is used for named placeholders. If multiple `Named`
//     maps are passed in args, the last one silently wins; merge them into one map
//     beforehand if you need keys from several sources.
//
// Security:
//   - Placeholders resolve data only: map keys, MapOrd keys, slice/array indices,
//     and struct fields. Methods are never invoked, so a template cannot execute
//     code on the supplied arguments.
//
// Usage:
//
//	// Example 1: Numeric placeholders
//	result := g.Format("{1} + {2} = {3}", 1, 2, 3)
//
//	// Example 2: Named placeholders
//	named := g.Named{
//		"name": "Alice",
//		"age":  30,
//	}
//	result := g.Format("My name is {name} and I am {age} years old.", named)
//
//	// Example 3: Field/key access
//	type User struct{ Name g.String }
//	user := User{Name: "Alice"}
//	result := g.Format("Name: {1.Name}", user)
//
//	// Example 4: Fallbacks
//	named := g.Named{
//		"name": g.String("John"),
//		"city": g.String("New York"),
//	}
//	result := g.Format("Hello, {name}. Welcome to {city?Unknown}!", named)
func Format[T ~string](template T, args ...any) String {
	return formatTemplate(template, args, "")
}

// FormatTo formats template and appends the result to builder without resetting it.
// It is intended for allocation-sensitive code that reuses a Builder across calls.
func FormatTo[T ~string](builder *Builder, template T, args ...any) {
	named, positional := formatArgs(args)
	parseTmplInto(builder, String(template), named, positional, nil, "")
}

// TryFormat validates template structure and argument resolution before
// formatting. Unlike Format, it returns an error for unmatched braces,
// missing values, malformed modifiers, and unsupported format verbs.
func TryFormat[T ~string](template T, args ...any) (result Result[String]) {
	defer func() {
		if recovered := recover(); recovered != nil {
			result = Err[String](fmt.Errorf("format: panic: %v", recovered))
		}
	}()

	tmpl := String(template)
	named, positional := formatArgs(args)
	if err := validateFormatTemplate(tmpl, named, positional); err != nil {
		return Err[String](err)
	}

	return Ok(parseTmpl(tmpl, named, positional, nil))
}

// TryFormatTo validates and formats into a temporary buffer, appending to
// builder only on success so an error never leaves a partial result behind.
func TryFormatTo[T ~string](builder *Builder, template T, args ...any) Result[Unit] {
	result := TryFormat(template, args...)
	if result.IsErr() {
		return Err[Unit](result.Err())
	}

	builder.WriteString(result.Ok())
	return Ok(Unit{})
}

func validateFormatTemplate(tmpl String, named Named, positional Slice[any]) error {
	length := tmpl.Len()
	var autoidx, idx Int

	for idx < length {
		char := tmpl[idx]
		if idx+1 < length && ((char == '{' && tmpl[idx+1] == '{') ||
			(char == '}' && tmpl[idx+1] == '}')) {
			idx += 2
			continue
		}

		if char == '}' {
			return fmt.Errorf("format: unmatched closing brace at byte %d", idx)
		}
		if char != '{' {
			idx++
			continue
		}

		cidx := tmpl[idx+1:].Index("}")
		if cidx.IsNegative() {
			return fmt.Errorf("format: unmatched opening brace at byte %d", idx)
		}

		eidx := idx + 1 + cidx
		placeholder, spec := splitFmtSpec(tmpl[idx+1 : eidx])

		trimmed := placeholder.Trim()
		if trimmed.IsEmpty() || trimmed[0] == '.' {
			autoidx++
			if autoidx > positional.Len() {
				return fmt.Errorf("format: missing automatic argument %d", autoidx)
			}
			if _, custom := positional[autoidx-1].(Formattable); !custom && !spec.IsEmpty() && !validFmtSpec(spec) {
				return fmt.Errorf("format: invalid format specifier %q", spec)
			}
			if !trimmed.IsEmpty() && !validModifierChain(trimmed[1:]) {
				return fmt.Errorf("format: malformed modifier chain %q", trimmed[1:])
			}
			idx = eidx + 1
			continue
		}

		keyfall, mods := placeholder, String("")
		if dot := placeholder.Index("."); !dot.IsNegative() {
			keyfall, mods = placeholder[:dot], placeholder[dot+1:]
		}
		key, fall := keyfall, String("")
		if q := keyfall.Index("?"); !q.IsNegative() {
			key, fall = keyfall[:q], keyfall[q+1:]
		}
		value := resolveValue(key, fall, named, positional)
		if value == nil {
			return fmt.Errorf("format: unresolved placeholder %q", placeholder)
		}
		if _, custom := value.(Formattable); !custom && !spec.IsEmpty() && !validFmtSpec(spec) {
			return fmt.Errorf("format: invalid format specifier %q", spec)
		}
		if !validModifierChain(mods) {
			return fmt.Errorf("format: malformed modifier chain %q", mods)
		}

		idx = eidx + 1
	}

	return nil
}

func validModifierChain(mods String) bool {
	valid := true
	forEachMod(mods, func(segment String) {
		if !valid {
			return
		}
		open := segment.Index("(")
		close := segment.LastIndex(")")
		if open.IsNegative() != close.IsNegative() || (!open.IsNegative() && close != segment.Len()-1) {
			valid = false
		}
	})

	return valid
}

func formatTemplate[T ~string](template T, args []any, suffix String) String {
	tmpl := String(template)
	named, positional := formatArgs(args)

	return parseTmplSuffix(tmpl, named, positional, nil, suffix)
}

// formatArgs separates the optional Named argument from positional arguments.
// The overwhelmingly common case contains neither Named nor nil, so reuse the
// caller-provided variadic slice instead of allocating and copying it.
func formatArgs(args []any) (Named, Slice[any]) {
	var named Named

	needsCopy, positionalLen := false, 0
	for _, arg := range args {
		switch x := arg.(type) {
		case Named:
			named = x
			needsCopy = true
		case nil:
			needsCopy = true
			positionalLen++
		default:
			positionalLen++
		}
	}

	if !needsCopy {
		return named, Slice[any](args)
	}

	if positionalLen == 0 {
		return named, nil
	}

	positional := make(Slice[any], 0, positionalLen)
	for _, arg := range args {
		switch x := arg.(type) {
		case Named:
			// Named arguments are metadata, not positional values. The last one
			// wins, matching the existing public contract.
		case nil:
			positional = append(positional, "<nil>")
		default:
			positional = append(positional, x)
		}
	}

	return named, positional
}

func parseTmpl(tmpl String, named Named, positional Slice[any], wraps *[]error) String {
	return parseTmplSuffix(tmpl, named, positional, wraps, "")
}

func parseTmplSuffix(tmpl String, named Named, positional Slice[any], wraps *[]error, suffix String) String {
	var builder Builder
	parseTmplInto(&builder, tmpl, named, positional, wraps, suffix)
	return builder.String()
}

func parseTmplInto(builder *Builder, tmpl String, named Named, positional Slice[any], wraps *[]error, suffix String) {
	length := tmpl.Len()
	builder.Grow(length + suffix.Len())

	var autoidx, idx Int

	for idx < length {
		char := tmpl[idx]
		if idx+1 < length && ((char == '{' && tmpl[idx+1] == '{') ||
			(char == '}' && tmpl[idx+1] == '}')) {
			builder.WriteByte(char)
			idx += 2
			continue
		}

		if char == '{' {
			cidx := tmpl[idx+1:].Index("}")
			if cidx.IsNegative() {
				builder.WriteByte(char)
				idx++

				continue
			}

			eidx := idx + 1 + cidx
			placeholder := tmpl[idx+1 : eidx]

			// extract format spec before auto-index check
			var fmtSuffix String
			if ci := findUnparenColon(placeholder); ci >= 0 {
				fmtSuffix = placeholder[ci:] // includes ':'
				placeholder = placeholder[:ci]
			}

			trimmed := placeholder.Trim()
			if trimmed.IsEmpty() || trimmed[0] == '.' {
				autoidx++
				if autoidx <= positional.Len() {
					mods := trimmed
					if !mods.IsEmpty() {
						mods = mods[1:]
					}

					formatSpec := fmtSuffix
					if !formatSpec.IsEmpty() {
						formatSpec = formatSpec[1:]
					}

					value := positional[autoidx-1]
					if mods.IsEmpty() && formatSpec.IsEmpty() {
						writeFormatValue(builder, value)
					} else if !mods.IsEmpty() ||
						!tryAppendNativeSpec(builder, value, parseFmtSpec(formatSpec)) {
						builder.WriteString(formatResolved(value, mods, formatSpec, wraps))
					}

					idx = eidx + 1
					continue
				}
			}

			// re-attach format spec
			if !fmtSuffix.IsEmpty() {
				placeholder += fmtSuffix
			}

			if fmtSuffix.IsEmpty() && placeholder.Index(".").IsNegative() && placeholder.Index("?").IsNegative() {
				if value := resolveValue(placeholder, "", named, positional); value != nil {
					writeFormatValue(builder, value)
					idx = eidx + 1
					continue
				}
			}

			replaced := processPlaceholder(placeholder, named, positional, wraps)
			builder.WriteString(replaced)

			idx = eidx + 1
		} else {
			builder.WriteByte(tmpl[idx])
			idx++
		}
	}

	builder.WriteString(suffix)
}

func writeFormatValue(builder *Builder, value any) {
	switch v := value.(type) {
	case Formattable:
		builder.WriteString(v.FormatValue(""))
	case string:
		builder.WriteString(String(v))
	case String:
		builder.WriteString(v)
	default:
		builder.WriteString(String(fmt.Sprint(value)))
	}
}

func processPlaceholder(placeholder String, named Named, positional Slice[any], wraps *[]error) String {
	// split off format spec
	placeholder, formatSpec := splitFmtSpec(placeholder)

	var (
		keyfall String
		mods    String
		key     String
		fall    String
	)

	if idx := placeholder.Index("."); !idx.IsNegative() {
		keyfall = placeholder[:idx]
		mods = placeholder[idx+1:]
	} else {
		keyfall = placeholder
	}

	if idx := keyfall.Index("?"); !idx.IsNegative() {
		key = keyfall[:idx]
		fall = keyfall[idx+1:]
	} else {
		key = keyfall
	}

	value := resolveValue(key, fall, named, positional)
	if value == nil {
		return "{" + placeholder + "}"
	}

	return formatResolved(value, mods, formatSpec, wraps)
}

func formatResolved(value any, mods, formatSpec String, wraps *[]error) String {
	if !mods.IsEmpty() {
		forEachMod(mods, func(segment String) {
			name, _ := parseMod(segment)
			value = applyMod(value, name)
		})
	}

	if formattable, ok := value.(Formattable); ok {
		return formattable.FormatValue(formatSpec)
	}

	if !formatSpec.IsEmpty() {
		spec := parseFmtSpec(formatSpec)
		if spec.verb == 'w' {
			if wraps != nil {
				if err, ok := value.(error); ok {
					*wraps = append(*wraps, err)
				}
			}

			return String(fmt.Sprint(value))
		}

		return applyFmtSpec(value, spec)
	}

	return String(fmt.Sprint(value))
}

// forEachMod scans a modifier chain without constructing an iterator, a slice
// of segments, or closures for Filter/ForEach.
func forEachMod(mods String, fn func(String)) {
	start := Int(0)
	for i := Int(0); i <= mods.Len(); i++ {
		if i < mods.Len() && mods[i] != '.' {
			continue
		}

		if segment := mods[start:i]; !segment.IsEmpty() {
			fn(segment)
		}

		start = i + 1
	}
}

func resolveValue(key, fall String, named Named, positional Slice[any]) any {
	if !key.IsEmpty() {
		first := key[0]
		if first >= '0' && first <= '9' || (first == '+' || first == '-') && key.Len() > 1 {
			if num := key.TryInt(); num.IsOk() {
				idx := num.v - 1
				if idx.IsNegative() || idx.Gte(positional.Len()) {
					return nil
				}

				return positional[idx]
			}
		}
	}

	value, ok := named[key]
	if !ok && !fall.IsEmpty() {
		value = named[fall]
	}

	return value
}

func parseMod(segment String) (String, Slice[String]) {
	oidx := segment.Index("(")
	if oidx.IsNegative() {
		return segment, nil
	}

	cidx := segment.LastIndex(")")
	if cidx.Lt(oidx) {
		return segment, nil
	}

	raw := segment[oidx+1 : cidx]
	count := 1
	for i := Int(0); i < raw.Len(); i++ {
		if raw[i] == ',' {
			count++
		}
	}

	params := make(Slice[String], 0, count)
	start := Int(0)
	for i := Int(0); i <= raw.Len(); i++ {
		if i < raw.Len() && raw[i] != ',' {
			continue
		}

		params = append(params, raw[start:i])
		start = i + 1
	}

	name := segment[:oidx]

	return name, params
}

func toType(param String, targetType reflect.Type) Result[reflect.Value] {
	switch targetType.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i64, err := strconv.ParseInt(param.Std(), 10, targetType.Bits())
		if err != nil {
			return Err[reflect.Value](err)
		}
		return Ok(reflect.ValueOf(i64).Convert(targetType))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u64, err := strconv.ParseUint(param.Std(), 10, targetType.Bits())
		if err != nil {
			return Err[reflect.Value](err)
		}
		return Ok(reflect.ValueOf(u64).Convert(targetType))
	case reflect.Bool:
		b, err := strconv.ParseBool(param.Std())
		if err != nil {
			return Err[reflect.Value](err)
		}
		return Ok(reflect.ValueOf(b))
	case reflect.Float32, reflect.Float64:
		param = param.ReplaceAll("_", ".")
		fl, err := strconv.ParseFloat(param.Std(), targetType.Bits())
		if err != nil {
			return Err[reflect.Value](err)
		}
		return Ok(reflect.ValueOf(fl).Convert(targetType))
	default:
		switch targetType {
		case reflect.TypeFor[string]():
			return Ok(reflect.ValueOf(param.Std()))
		case reflect.TypeFor[String]():
			return Ok(reflect.ValueOf(param))
		default:
			return Err[reflect.Value](fmt.Errorf("unsupported type: %s", targetType))
		}
	}
}

func extractFromMapOrd(param String, slice reflect.Value) Option[any] {
	for slice.Kind() == reflect.Interface || slice.Kind() == reflect.Pointer {
		if slice.IsNil() {
			return None[any]()
		}
		slice = slice.Elem()
	}

	if !slice.IsValid() || slice.Kind() != reflect.Slice {
		return None[any]()
	}

	elemT := slice.Type().Elem()
	if elemT.Kind() != reflect.Struct {
		return None[any]()
	}

	if _, ok := elemT.FieldByName("Key"); !ok {
		return None[any]()
	}

	if _, ok := elemT.FieldByName("Value"); !ok {
		return None[any]()
	}

	ps := param.Std()

	var (
		pi    int
		pierr error
		pf    float64
		pferr error
	)

	for i := 0; i < slice.Len(); i++ {
		el := slice.Index(i)
		k := el.FieldByName("Key").Interface()
		v := el.FieldByName("Value").Interface()

		switch kk := k.(type) {
		case string:
			if kk == ps {
				return Some(v)
			}
		case String:
			if string(kk) == ps {
				return Some(v)
			}
		case int:
			if pi == 0 && pierr == nil {
				pi, pierr = strconv.Atoi(ps)
			}
			if pierr == nil && kk == pi {
				return Some(v)
			}
		case Int:
			if pi == 0 && pierr == nil {
				pi, pierr = strconv.Atoi(ps)
			}
			if pierr == nil && int(kk) == pi {
				return Some(v)
			}
		case float64:
			if pf == 0 && pferr == nil {
				pf, pferr = strconv.ParseFloat(param.ReplaceAll("_", ".").Std(), 64)
			}
			if pferr == nil && kk == pf {
				return Some(v)
			}
		case Float:
			if pf == 0 && pferr == nil {
				pf, pferr = strconv.ParseFloat(param.ReplaceAll("_", ".").Std(), 64)
			}
			if pferr == nil && float64(kk) == pf {
				return Some(v)
			}
		default:
			if s, ok := k.(fmt.Stringer); ok && s.String() == ps {
				return Some(v)
			}
		}
	}

	return None[any]()
}

func resolveIndirect(targetType reflect.Value) reflect.Value {
	for targetType.Kind() == reflect.Interface || targetType.Kind() == reflect.Pointer {
		if targetType.IsNil() {
			return reflect.Value{}
		}

		targetType = targetType.Elem()
	}

	return targetType
}

// applyMod resolves one dot-segment of a placeholder as DATA access only: a
// map key, a MapOrd key, a slice/array index, or a struct field.
//
// It deliberately does NOT invoke methods. Method modifiers ({x.Trim.Upper})
// were removed: they silently broke whenever a method moved out of the root
// package, and their dynamic MethodByName call disabled the linker's
// dead-method elimination for every binary linking g, bloating binaries far
// beyond this package. Transform the value before handing it to Format.
func applyMod(value any, name String) any {
	current := reflect.ValueOf(value)

	for current.Kind() == reflect.Pointer || current.Kind() == reflect.Interface {
		if current.IsNil() {
			return value
		}
		current = current.Elem()
	}

	switch current.Kind() {
	case reflect.Map:
		key := toType(name, current.Type().Key())
		if key.IsErr() {
			return value
		}

		current = resolveIndirect(current.MapIndex(key.v))
	case reflect.Slice, reflect.Array:
		if pair := extractFromMapOrd(name, current); pair.IsSome() {
			return pair.v
		}

		idx := name.TryInt()
		if idx.IsErr() || idx.v.Gte(Int(current.Len())) {
			return value
		}

		current = current.Index(idx.v.Std())
	case reflect.Struct:
		current = current.FieldByName(name.Std())
	}

	if current.IsValid() && current.CanInterface() {
		return current.Interface()
	}

	return value
}
