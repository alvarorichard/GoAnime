package g

import "strings"

// Builder wraps strings.Builder and provides additional type-safe methods
// for use with the custom types String and Int.
type Builder struct{ builder strings.Builder }

// NewBuilder creates a new instance of Builder.
func NewBuilder() *Builder { return new(Builder) }

// Write appends the given byte slice to the builder.
func (b *Builder) Write(bs []byte) (int, error) { return b.builder.Write(bs) }

// WriteString appends the given String to the builder.
func (b *Builder) WriteString(str String) (int, error) { return b.builder.WriteString(str.Std()) }

// WriteByte appends the given byte to the builder.
func (b *Builder) WriteByte(c byte) error { return b.builder.WriteByte(c) }

// WriteRune appends the given rune to the builder.
func (b *Builder) WriteRune(r rune) (int, error) { return b.builder.WriteRune(r) }

// Grow increases the builder’s capacity by at least n bytes.
func (b *Builder) Grow(n Int) { b.builder.Grow(n.Std()) }

// Cap returns the builder’s current capacity.
func (b *Builder) Cap() Int { return Int(b.builder.Cap()) }

// Len returns the number of bytes currently in the builder.
func (b *Builder) Len() Int { return Int(b.builder.Len()) }

// Reset clears the contents of the builder.
func (b *Builder) Reset() { b.builder.Reset() }

// String returns the accumulated string as a custom String type.
func (b *Builder) String() String { return String(b.builder.String()) }
