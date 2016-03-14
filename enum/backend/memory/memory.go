package memory

import (
	. "enum-dns/enum"
	"fmt"
	"sort"
)

type storage struct {
	entries []NumberRange
}

type Asc []NumberRange

func (a Asc) Len() int           { return len(a) }
func (a Asc) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a Asc) Less(i, j int) bool { return a[i].Lower < a[j].Lower }

type memoryBackend struct {
	s *storage
}

func NewMemoryBackend() (Backend, error) {
	// TODO: Investigate this black magic...
	return &memoryBackend{s: &storage{entries: make([]NumberRange, 0)}}, nil
}

func (b *memoryBackend) RangesBetween(l, u uint64, c int) ([]NumberRange, error) {
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

func (b *memoryBackend) PushRange(add NumberRange) ([]NumberRange, error) {
	// Just add
	l, err := PrefixToE164(add.Lower)
	if err != nil {
		return nil, err
	}
	u, err := PrefixToE164(add.Upper)
	if err != nil {
		return nil, err
	}

	results := make([]NumberRange, 0)
	for i := len(b.s.entries) - 1; i >= 0; i-- {
		entry := &b.s.entries[i]
		fmt.Printf("checking [%d:%d] against [%d:%d] \n", entry.Lower, entry.Upper, add.Lower, add.Upper)
		if add.Contains(*entry) {
			b.s.entries = append(b.s.entries[:i], b.s.entries[i+1:]...)
			fmt.Printf("delete [%d:%d]\n", entry.Lower, entry.Upper)
			results = append(results, *entry)
		} else if entry.OverlapWith(add) {
			if entry.Lower <= add.Upper && add.Upper <= entry.Upper {
				fmt.Printf("adjust [%d:%d] to [%d:%d]\n", entry.Lower, entry.Upper, add.Upper+1, entry.Upper)
				entry.Lower = add.Upper + 1
			}
			if entry.Lower <= add.Lower && add.Lower <= entry.Upper {
				fmt.Printf("adjust [%d:%d] to [%d:%d]\n", entry.Lower, entry.Upper, entry.Lower, add.Lower-1)
				entry.Upper = add.Lower - 1
			}
		}
	}

	add.Lower = l
	add.Upper = u
	fmt.Printf("adding [%d:%d]\n", add.Lower, add.Upper)
	b.s.entries = append(b.s.entries, add)
	sort.Sort(Asc(b.s.entries))

	for _, v := range b.s.entries {
		fmt.Printf("end result: [%d:%d]\n", v.Lower, v.Upper)
	}

	return results, nil
}

func (b *memoryBackend) Close() error {
	return nil
}
