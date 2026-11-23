// Package godebug reads the per-package compatibility settings used by this fork.
// It deliberately does not register counters or hooks in the Go runtime.
package godebug

import (
	"os"
	"strings"
)

type Setting struct{ name string }

func New(name string) *Setting { return &Setting{name: name} }
func (s *Setting) Value() string {
	var value string
	for field := range strings.SplitSeq(os.Getenv("GODEBUG"), ",") {
		if name, v, ok := strings.Cut(field, "="); ok && name == s.name {
			value = v
		}
	}
	return value
}
func (*Setting) IncNonDefault() {}
