package cmp

import "fmt"

// Ordering is the result of a comparison between two values.
type Ordering int

const (
	Less    Ordering = iota - 1 // Less represents an ordered value where a compared value is less than another.
	Equal                       // Equal represents an ordered value where a compared value is equal to another.
	Greater                     // Greater represents an ordered value where a compared value is greater than another.
)

// Then returns other if the receiver is Equal, otherwise returns the receiver.
// It is useful for chaining comparisons, where the first non-Equal result wins.
func (o Ordering) Then(other Ordering) Ordering {
	if o.IsEq() {
		return other
	}

	return o
}

// Reverse reverses the ordering.
func (o Ordering) Reverse() Ordering {
	switch o {
	case Less:
		return Greater
	case Greater:
		return Less
	default:
		return Equal
	}
}

// IsLt returns true if the Ordering value is Less.
func (o Ordering) IsLt() bool { return Less == o }

// IsEq returns true if the Ordering value is Equal.
func (o Ordering) IsEq() bool { return Equal == o }

// IsGt returns true if the Ordering value is Greater.
func (o Ordering) IsGt() bool { return Greater == o }

// String returns the string representation of the Ordering value.
func (o Ordering) String() string {
	switch o {
	case Less:
		return "Less"
	case Equal:
		return "Equal"
	case Greater:
		return "Greater"
	default:
		return fmt.Sprintf("Unknown Ordering value: %d", int(o))
	}
}
