package memory

import (
	. "enum-dns/enum"
	"fmt"
)

type storage struct {
	entries []NumberRange
}
type memoryBackend struct {
	s *storage
}

func NewMemoryBackend() (Backend, error) {
	// TODO: Investigate this black magic...
	return &memoryBackend{s: &storage{entries: make([]NumberRange, 0)}}, nil
}

func (b memoryBackend) RangesBetween(l, u uint64, c int) ([]NumberRange, error) {
	results := make([]NumberRange, 0)
	r := NumberRange{Lower: l, Upper: u}
	switch {
	case c < 0:
		for i := len(b.s.entries) - 1; i >= 0 && c != 0; i-- {
			if entry := b.s.entries[i]; entry.OverlapWith(r) {
				results = append(results, entry)
				c++
			}
		}
	case c > 0:
		for i := 0; i < len(b.s.entries) && c != 0; i++ {
			if entry := b.s.entries[i]; entry.OverlapWith(r) {
				results = append(results, entry)
				c--
			}
		}
	}
	return results, nil
}

func (b memoryBackend) PushRange(add NumberRange) ([]NumberRange, error) {
	// Just add
	l, err := PrefixToE164(add.Lower)
	if err != nil {
		return nil, err
	}
	u, err := PrefixToE164(add.Upper)
	if err != nil {
		return nil, err
	}

	add.Lower = l
	add.Upper = u
	fmt.Printf("[%d:%d]", add.Lower, add.Upper)
	b.s.entries = append(b.s.entries, add)
	return nil, nil
}

func (b memoryBackend) Close() error {
	return nil
}
