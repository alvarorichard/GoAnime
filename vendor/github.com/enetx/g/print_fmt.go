package g

import (
	"fmt"
	"math"
	"strconv"
	"unicode/utf8"
)

type fmtAlign int

const (
	alignNone   fmtAlign = iota
	alignLeft            // '<'
	alignRight           // '>'
	alignCenter          // '^'
)

type fmtSign int

const (
	signNone  fmtSign = iota
	signPlus          // '+'
	signMinus         // '-'
	signSpace         // ' '
)

type fmtSpec struct {
	fill      rune
	align     fmtAlign
	sign      fmtSign
	alternate bool
	zeroPad   bool
	width     Int
	precision Int
	verb      byte
}

func splitFmtSpec(placeholder String) (String, String) {
	if i := findUnparenColon(placeholder); i >= 0 {
		return placeholder[:i], placeholder[i+1:]
	}

	return placeholder, ""
}

func findUnparenColon(placeholder String) Int {
	depth := 0

	for i := Int(0); i < placeholder.Len(); i++ {
		switch placeholder[i] {
		case '(':
			depth++
		case ')':
			depth--
		case ':':
			if depth == 0 {
				return i
			}
		}
	}

	return -1
}

// maxFmtDigits caps the accumulated width/precision value to avoid huge
// allocations or integer overflow on absurd inputs like {:99999999999999999999}.
// 1<<20 is far larger than any reasonable field width while staying well below
// int range on every platform.
const maxFmtDigits = 1 << 20

func parseDigits(runes []rune, pos int) (int, int, bool) {
	n, has := 0, false

	for pos < len(runes) && runes[pos] >= '0' && runes[pos] <= '9' {
		if n <= maxFmtDigits {
			n = min(n*10+int(runes[pos]-'0'), maxFmtDigits)
		}

		pos++
		has = true
	}

	return n, pos, has
}

func validFmtSpec(spec String) bool {
	runes := []rune(spec)
	pos := 0

	if pos < len(runes) {
		if pos+1 < len(runes) && isAlign(runes[pos+1]) {
			pos += 2
		} else if isAlign(runes[pos]) {
			pos++
		}
	}

	if pos < len(runes) && (runes[pos] == '+' || runes[pos] == '-' || runes[pos] == ' ') {
		pos++
	}

	if pos < len(runes) && runes[pos] == '#' {
		pos++
	}

	if pos < len(runes) && runes[pos] == '0' {
		pos++
	}

	_, pos, _ = parseDigits(runes, pos)
	if pos < len(runes) && runes[pos] == '.' {
		pos++
		_, pos, _ = parseDigits(runes, pos)
	}

	if pos == len(runes) {
		return true
	}

	verb := runes[pos]

	pos++
	if pos != len(runes) {
		return false
	}

	switch verb {
	case 'd', 'c', 'q', 'U', 'x', 'X', 'o', 'b', 'e', 'E', 'T', 'p', '?', 'w', 's', 'f':
		return true
	default:
		return false
	}
}

func parseFmtSpec(spec String) fmtSpec {
	if isASCIIFmtSpec(spec) {
		return parseASCIIFmtSpec(spec)
	}

	return parseRuneFmtSpec(spec)
}

func isASCIIFmtSpec(spec String) bool {
	for i := Int(0); i < spec.Len(); i++ {
		if spec[i] >= utf8.RuneSelf {
			return false
		}
	}

	return true
}

func parseASCIIDigits(spec String, pos Int) (int, Int, bool) {
	n, has := 0, false
	for pos < spec.Len() && spec[pos] >= '0' && spec[pos] <= '9' {
		if n <= maxFmtDigits {
			n = min(n*10+int(spec[pos]-'0'), maxFmtDigits)
		}

		pos++
		has = true
	}

	return n, pos, has
}

func parseASCIIFmtSpec(spec String) fmtSpec {
	fs := fmtSpec{fill: ' ', width: -1, precision: -1}
	pos := Int(0)

	if pos < spec.Len() {
		if pos+1 < spec.Len() && isAlign(rune(spec[pos+1])) {
			fs.fill = rune(spec[pos])
			fs.align = toAlign(rune(spec[pos+1]))
			pos += 2
		} else if isAlign(rune(spec[pos])) {
			fs.align = toAlign(rune(spec[pos]))
			pos++
		}
	}

	if pos < spec.Len() {
		switch spec[pos] {
		case '+':
			fs.sign = signPlus
			pos++
		case '-':
			fs.sign = signMinus
			pos++
		case ' ':
			fs.sign = signSpace
			pos++
		}
	}

	if pos < spec.Len() && spec[pos] == '#' {
		fs.alternate = true
		pos++
	}

	if pos < spec.Len() && spec[pos] == '0' && fs.align == alignNone {
		fs.zeroPad = true
		pos++
	}

	if n, next, ok := parseASCIIDigits(spec, pos); ok {
		fs.width, pos = Int(n), next
	}

	if pos < spec.Len() && spec[pos] == '.' {
		pos++
		if n, next, ok := parseASCIIDigits(spec, pos); ok {
			fs.precision, pos = Int(n), next
		} else {
			fs.precision = 0
		}
	}

	if pos < spec.Len() {
		fs.verb = spec[pos]
	}

	return fs
}

func parseRuneFmtSpec(spec String) fmtSpec {
	fs := fmtSpec{fill: ' ', width: -1, precision: -1}

	runes := []rune(spec)
	pos := 0

	// parse [[fill]align]
	if pos < len(runes) {
		if pos+1 < len(runes) && isAlign(runes[pos+1]) {
			fs.fill = runes[pos]
			fs.align = toAlign(runes[pos+1])
			pos += 2
		} else if isAlign(runes[pos]) {
			fs.align = toAlign(runes[pos])
			pos++
		}
	}

	// parse [sign]
	if pos < len(runes) {
		switch runes[pos] {
		case '+':
			fs.sign = signPlus
			pos++
		case '-':
			fs.sign = signMinus
			pos++
		case ' ':
			fs.sign = signSpace
			pos++
		}
	}

	// parse [#]
	if pos < len(runes) && runes[pos] == '#' {
		fs.alternate = true
		pos++
	}

	// parse [0]
	if pos < len(runes) && runes[pos] == '0' && fs.align == alignNone {
		fs.zeroPad = true
		pos++
	}

	// parse [width]
	if n, np, ok := parseDigits(runes, pos); ok {
		fs.width = Int(n)
		pos = np
	}

	// parse [.precision]
	if pos < len(runes) && runes[pos] == '.' {
		pos++

		if n, np, ok := parseDigits(runes, pos); ok {
			fs.precision = Int(n)
			pos = np
		} else {
			fs.precision = 0
		}
	}

	// parse [type]
	if pos < len(runes) {
		fs.verb = byte(runes[pos])
	}

	return fs
}

func isAlign(r rune) bool { return r == '<' || r == '>' || r == '^' }

func toAlign(r rune) fmtAlign {
	switch r {
	case '<':
		return alignLeft
	case '>':
		return alignRight
	default:
		return alignCenter
	}
}

func applyFmtSpec(value any, spec fmtSpec) String {
	var s String

	switch spec.verb {
	case 'd':
		s = fmtDecimal(value, spec)
	case 'c':
		s = fmtCharacter(value)
	case 'q':
		s = fmtQuoted(value)
	case 'U':
		s = fmtUnicode(value, spec.alternate)
	case 'x':
		s = fmtIntBase(value, 16, fmtAltPrefix("0x", spec), spec, false)
	case 'X':
		s = fmtIntBase(value, 16, fmtAltPrefix("0x", spec), spec, true)
	case 'o':
		s = fmtIntBase(value, 8, fmtAltPrefix("0o", spec), spec, false)
	case 'b':
		s = fmtIntBase(value, 2, fmtAltPrefix("0b", spec), spec, false)
	case 'e':
		s = fmtExponential(value, spec, 'e')
	case 'E':
		s = fmtExponential(value, spec, 'E')
	case 'T':
		s = String(fmt.Sprintf("%T", value))
	case 'p':
		s = String(fmt.Sprintf("%p", value))
	case '?':
		if spec.alternate {
			s = fmtPrettyDebug(value)
		} else {
			s = String(fmt.Sprintf("%#v", value))
		}
	default:
		s = fmtDefault(value, spec)
	}

	return fmtPad(s, spec, fmtIsNumeric(value))
}

func tryAppendNativeSpec(builder *Builder, value any, spec fmtSpec) bool {
	return tryAppendNativeIntSpec(builder, value, spec) ||
		tryAppendNativeTextSpec(builder, value, spec)
}

// tryAppendNativeIntSpec handles the common integer representations directly in
// the destination builder. It avoids the temporary strings otherwise produced
// by strconv.Format*, sign/prefix concatenation, and padding.
func tryAppendNativeIntSpec(builder *Builder, value any, spec fmtSpec) bool {
	base := 10
	prefix := String("")
	upper := false

	switch spec.verb {
	case 'd':
	case 'x':
		base = 16
		if spec.alternate {
			prefix = "0x"
		}
	case 'X':
		base, upper = 16, true
		if spec.alternate {
			prefix = "0x"
		}
	case 'o':
		base = 8
		if spec.alternate {
			prefix = "0o"
		}
	case 'b':
		base = 2
		if spec.alternate {
			prefix = "0b"
		}
	default:
		return false
	}

	var storage [128]byte
	digits := storage[:0]
	negative := false
	if u, ok := fmtToUint64(value); ok {
		digits = strconv.AppendUint(digits, u, base)
	} else if i, ok := fmtToInt64(value); ok {
		negative = i < 0
		if negative {
			// -(i+1)+1 is safe for MinInt64.
			digits = strconv.AppendUint(digits, uint64(-(i+1))+1, base)
		} else {
			digits = strconv.AppendUint(digits, uint64(i), base)
		}
	} else {
		return false
	}

	if upper {
		for i, ch := range digits {
			if ch >= 'a' && ch <= 'f' {
				digits[i] = ch - ('a' - 'A')
			}
		}
	}

	var sign byte
	if negative {
		sign = '-'
	} else if spec.sign == signPlus {
		sign = '+'
	} else if spec.sign == signSpace {
		sign = ' '
	}

	contentLen := Int(len(digits)) + prefix.Len()
	if sign != 0 {
		contentLen++
	}
	padding := Int(0)
	if spec.width > contentLen {
		padding = spec.width - contentLen
	}

	if spec.zeroPad && spec.align == alignNone {
		if sign != 0 {
			builder.WriteByte(sign)
		}
		builder.WriteString(prefix)
		for range padding {
			builder.WriteByte('0')
		}
		builder.Write(digits)
		return true
	}

	align := spec.align
	if align == alignNone {
		align = alignRight
	}

	left, right := Int(0), Int(0)
	switch align {
	case alignLeft:
		right = padding
	case alignRight:
		left = padding
	default:
		left = padding / 2
		right = padding - left
	}

	for range left {
		builder.WriteRune(spec.fill)
	}

	if sign != 0 {
		builder.WriteByte(sign)
	}

	builder.WriteString(prefix)
	builder.Write(digits)

	for range right {
		builder.WriteRune(spec.fill)
	}

	return true
}

// tryAppendNativeTextSpec writes character, quoted, and Unicode representations
// without allocating intermediate strings.
func tryAppendNativeTextSpec(builder *Builder, value any, spec fmtSpec) bool {
	var storage [128]byte
	content := storage[:0]

	switch spec.verb {
	case 'c':
		r, ok := fmtToRune(value)
		if !ok {
			return false
		}
		content = utf8.AppendRune(content, r)
	case 'q':
		switch v := value.(type) {
		case string:
			content = strconv.AppendQuote(content, v)
		case String:
			content = strconv.AppendQuote(content, v.Std())
		case []byte:
			content = strconv.AppendQuote(content, string(v))
		case Bytes:
			content = strconv.AppendQuote(content, string(v))
		default:
			r, ok := fmtToRune(value)
			if !ok {
				return false
			}
			content = strconv.AppendQuoteRune(content, r)
		}
	case 'U':
		r, ok := fmtToRune(value)
		if !ok || !utf8.ValidRune(r) {
			return false
		}
		content = append(content, 'U', '+')
		var digits [8]byte
		hex := strconv.AppendInt(digits[:0], int64(r), 16)
		for padding := 4 - len(hex); padding > 0; padding-- {
			content = append(content, '0')
		}
		for _, ch := range hex {
			if ch >= 'a' && ch <= 'f' {
				ch -= 'a' - 'A'
			}
			content = append(content, ch)
		}
		if spec.alternate && strconv.IsPrint(r) {
			content = append(content, ' ')
			content = strconv.AppendQuoteRune(content, r)
		}
	default:
		return false
	}

	appendNativePadded(builder, content, utf8.RuneCount(content), spec, fmtIsNumeric(value))
	return true
}

func fmtToRune(value any) (rune, bool) {
	if u, ok := fmtToUint64(value); ok {
		return rune(u), true
	}
	if i, ok := fmtToSignedInt64(value); ok {
		return rune(i), true
	}
	return 0, false
}

func appendNativePadded(builder *Builder, content []byte, runes int, spec fmtSpec, isNum bool) {
	padding := spec.width - Int(runes)
	if padding <= 0 {
		builder.Write(content)
		return
	}

	fill := spec.fill
	align := spec.align
	if spec.zeroPad && align == alignNone && isNum {
		fill, align = '0', alignRight
	} else if align == alignNone {
		if isNum {
			align = alignRight
		} else {
			align = alignLeft
		}
	}

	left, right := Int(0), Int(0)
	switch align {
	case alignLeft:
		right = padding
	case alignRight:
		left = padding
	default:
		left = padding / 2
		right = padding - left
	}

	for range left {
		builder.WriteRune(fill)
	}
	builder.Write(content)
	for range right {
		builder.WriteRune(fill)
	}
}

func fmtDecimal(value any, spec fmtSpec) String {
	if u, ok := fmtToUint64(value); ok {
		return fmtApplySign(String(strconv.FormatUint(u, 10)), true, spec)
	}

	if i, ok := fmtToSignedInt64(value); ok {
		s := String(strconv.FormatInt(i, 10))
		return fmtApplySign(s, i >= 0, spec)
	}

	return String(fmt.Sprint(value))
}

func fmtCharacter(value any) String {
	if u, ok := fmtToUint64(value); ok {
		return String(string(rune(u)))
	}

	if i, ok := fmtToSignedInt64(value); ok {
		return String(string(rune(i)))
	}

	return String(fmt.Sprintf("%c", value))
}

func fmtQuoted(value any) String {
	switch v := value.(type) {
	case string:
		return String(strconv.Quote(v))
	case String:
		return String(strconv.Quote(v.Std()))
	case []byte:
		return String(strconv.Quote(string(v)))
	case Bytes:
		return String(strconv.Quote(string(v)))
	}

	if u, ok := fmtToUint64(value); ok {
		return String(strconv.QuoteRune(rune(u)))
	}

	if i, ok := fmtToSignedInt64(value); ok {
		return String(strconv.QuoteRune(rune(i)))
	}

	return String(fmt.Sprintf("%q", value))
}

func fmtUnicode(value any, alternate bool) String {
	var r rune
	if u, ok := fmtToUint64(value); ok {
		r = rune(u)
	} else if i, ok := fmtToSignedInt64(value); ok {
		r = rune(i)
	} else {
		verb := "%U"
		if alternate {
			verb = "%#U"
		}
		return String(fmt.Sprintf(verb, value))
	}

	if alternate {
		return String(fmt.Sprintf("%#U", r))
	}

	return String(fmt.Sprintf("%U", r))
}

func fmtAltPrefix(prefix String, spec fmtSpec) String {
	if spec.alternate {
		return prefix
	}

	return ""
}

func fmtIntBase(value any, base int, prefix String, spec fmtSpec, upper bool) String {
	format := func(s String) String {
		if upper {
			return prefix + s.Upper()
		}

		return prefix + s
	}

	if u, ok := fmtToUint64(value); ok {
		return format(String(strconv.FormatUint(u, base)))
	}

	if i, ok := fmtToInt64(value); ok {
		s := format(String(strconv.FormatInt(abs64(i), base)))

		if i < 0 {
			return "-" + s
		}

		return fmtApplySign(s, true, spec)
	}

	return String(fmt.Sprint(value))
}

func fmtDefault(value any, spec fmtSpec) String {
	prec := -1
	if spec.precision >= 0 {
		prec = spec.precision.Std()
	}

	if prec >= 0 || spec.sign != signNone {
		if f, ok := fmtToFloat64(value); ok {
			return fmtApplySign(
				String(strconv.FormatFloat(f, 'f', prec, 64)), f >= 0, spec,
			)
		}

		if spec.sign != signNone {
			if i, ok := fmtToInt64(value); ok {
				return fmtApplySign(
					String(strconv.FormatInt(abs64(i), 10)), i >= 0, spec,
				)
			}

			// uint64 values above MaxInt64 don't fit in int64; format them
			// unsigned so the requested sign (e.g. {:+}) is still applied.
			if u, ok := fmtToUint64(value); ok {
				return fmtApplySign(
					String(strconv.FormatUint(u, 10)), true, spec,
				)
			}
		}
	}

	s := String(fmt.Sprint(value))
	if prec >= 0 && s.LenRunes() > spec.precision {
		return String([]rune(s)[:spec.precision])
	}

	return s
}

func fmtExponential(value any, spec fmtSpec, verb byte) String {
	prec := 6
	if spec.precision >= 0 {
		prec = spec.precision.Std()
	}

	f, ok := fmtToFloat64(value)
	if !ok {
		var i int64
		if i, ok = fmtToInt64(value); ok {
			f = float64(i)
		}
	}

	if ok {
		return fmtApplySign(
			String(strconv.FormatFloat(f, verb, prec, 64)), f >= 0, spec,
		)
	}

	return String(fmt.Sprint(value))
}

func fmtApplySign(s String, positive bool, spec fmtSpec) String {
	if !positive {
		if s[0] != '-' {
			return "-" + s
		}

		return s
	}

	switch spec.sign {
	case signPlus:
		return "+" + s
	case signSpace:
		return " " + s
	}

	return s
}

func fmtPad(s String, spec fmtSpec, isNum bool) String {
	if spec.width.IsNegative() || s.LenRunes() >= spec.width {
		return s
	}

	// zero-pad: sign-aware
	if spec.zeroPad && spec.align == alignNone && isNum {
		return fmtZeroPad(s, spec.width)
	}

	// default alignment: right for numbers, left for strings
	align := spec.align
	if align == alignNone {
		if isNum {
			align = alignRight
		} else {
			align = alignLeft
		}
	}

	return fmtJustify(s, spec.width, spec.fill, align)
}

func fmtJustify(s String, width Int, fill rune, align fmtAlign) String {
	remaining := width - s.LenRunes()
	left, right := Int(0), Int(0)

	switch align {
	case alignLeft:
		right = remaining
	case alignRight:
		left = remaining
	default:
		left = remaining / 2
		right = remaining - left
	}

	var b Builder
	b.Grow(s.Len() + remaining*Int(utf8.RuneLen(fill)))

	for range left {
		b.WriteRune(fill)
	}

	b.WriteString(s)

	for range right {
		b.WriteRune(fill)
	}

	return b.String()
}

func fmtZeroPad(s String, width Int) String {
	var sign byte

	body := s
	if s.Len() > 0 && (s[0] == '+' || s[0] == '-' || s[0] == ' ') {
		sign = s[0]
		body = s[1:]
	}

	var prefix String
	if body.Len() > 1 && body[0] == '0' && (body[1] == 'x' || body[1] == 'X' || body[1] == 'b' || body[1] == 'o') {
		prefix = body[:2]
		body = body[2:]
	}

	padLen := width - body.LenRunes() - prefix.Len()
	if sign != 0 {
		padLen--
	}

	var b Builder
	b.Grow(width)

	if sign != 0 {
		b.WriteByte(sign)
	}

	b.WriteString(prefix)

	for range padLen {
		b.WriteByte('0')
	}

	b.WriteString(body)

	return b.String()
}

func fmtPrettyDebug(value any) String {
	return fmtIndentDebug(String(fmt.Sprintf("%#v", value)))
}

func fmtIndentDebug(s String) String {
	var b Builder
	b.Grow(s.Len() + 16)

	indent := Int(0)
	inString := false
	escaped := false
	pad := String("  ")

	for i := Int(0); i < s.Len(); i++ {
		ch := s[i]

		if inString {
			b.WriteByte(ch)

			// Track escapes with a running toggle so that a closing quote
			// preceded by an escaped backslash (`\\"`) is not mistaken for an
			// escaped quote. A single-char lookback would leave inString stuck.
			if escaped {
				escaped = false
			} else if ch == '\\' {
				escaped = true
			} else if ch == '"' {
				inString = false
			}

			continue
		}

		if ch == '"' {
			inString = true
			escaped = false
			b.WriteByte(ch)

			continue
		}

		switch ch {
		case '{', '[':
			indent++
			b.WriteByte(ch)
			b.WriteByte('\n')
			writeIndent(&b, pad, indent)
		case '}', ']':
			indent--
			b.WriteByte('\n')
			writeIndent(&b, pad, indent)
			b.WriteByte(ch)
		case ',':
			b.WriteByte(ch)
			b.WriteByte('\n')
			writeIndent(&b, pad, indent)
		default:
			b.WriteByte(ch)
		}
	}

	return b.String()
}

func writeIndent(b *Builder, pad String, indent Int) {
	for range indent {
		b.WriteString(pad)
	}
}

func fmtIsNumeric(value any) bool {
	switch value.(type) {
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64,
		Int, Float:
		return true
	}

	return false
}

func fmtToInt64(value any) (int64, bool) {
	switch v := value.(type) {
	case Int:
		return int64(v), true
	case int:
		return int64(v), true
	case int8:
		return int64(v), true
	case int16:
		return int64(v), true
	case int32:
		return int64(v), true
	case int64:
		return v, true
	case uint:
		return int64(v), true
	case uint8:
		return int64(v), true
	case uint16:
		return int64(v), true
	case uint32:
		return int64(v), true
	case uint64:
		if v <= math.MaxInt64 {
			return int64(v), true
		}

		return 0, false
	case Float:
		return int64(v), true
	case float32:
		return int64(v), true
	case float64:
		return int64(v), true
	}

	return 0, false
}

func fmtToSignedInt64(value any) (int64, bool) {
	switch v := value.(type) {
	case Int:
		return int64(v), true
	case int:
		return int64(v), true
	case int8:
		return int64(v), true
	case int16:
		return int64(v), true
	case int32:
		return int64(v), true
	case int64:
		return v, true
	}

	return 0, false
}

func fmtToUint64(value any) (uint64, bool) {
	switch v := value.(type) {
	case uint:
		return uint64(v), true
	case uint8:
		return uint64(v), true
	case uint16:
		return uint64(v), true
	case uint32:
		return uint64(v), true
	case uint64:
		return v, true
	}

	return 0, false
}

func fmtToFloat64(value any) (float64, bool) {
	switch v := value.(type) {
	case Float:
		return float64(v), true
	case float32:
		return float64(v), true
	case float64:
		return v, true
	}

	return 0, false
}

func abs64(i int64) int64 {
	if i < 0 {
		return -i
	}

	return i
}
