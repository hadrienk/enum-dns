// backend.go
package enum

import (
	"fmt"
)

type NumberRange struct {
	Upper       uint64
	Lower       uint64
	Order       uint16
	Preference  uint16
	Flags       string
	Service     string
	Regexp      string
	Replacement string
}

// RangeOverlapError is returned when an operation fails because
// a range overlaps with on or more other ranges.
type RangeOverlapError struct {
	Range    NumberRange
	Overlaps []NumberRange
}

func (e *RangeOverlapError) Error() string {
	if len(e.Overlaps) == 1 {
		return fmt.Sprintf("[%15.d:%15.d] orverlaps with [%15.d:%15.d]",
			e.Range.Lower, e.Range.Upper,
			e.Overlaps[0].Upper, e.Overlaps[0].Upper)
	} else {
		return fmt.Sprintf("[%15.d:%15.d] orverlaps with %d other ranges", e.Range.Lower, e.Range.Upper, len(e.Overlaps))
	}
}

type Backend interface {
	// Ranges returns a list of ranges. The n and c values allow to navigate.
	// n is the number to start from. c is the count of values to return. A negative
	// c value will return the values in reverse order.
	Ranges(n uint64, c int) ([]NumberRange, error)

	// RangeFor returns the NumberRange for a given n number, or nil if no NumberRange is
	// found.
	RangeFor(n uint64) (NumberRange, error)

	// Add a range to the backend. The operation can fail if the range overlap
	AddRange(r NumberRange) ([]NumberRange, error)

	// Remove a particular range from the backend. Returns an error if the deletion
	// fails.
	RemoveRange(r NumberRange) error

	// Close the backend.
	Close() error
}