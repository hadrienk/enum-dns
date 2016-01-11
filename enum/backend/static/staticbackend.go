// staticbackend.go
package enum

import . "enum-dns/enum"

type backend struct{}

func NewStaticBackend() Backend {
	return backend{}
}

func (b backend) Close() error {
	return nil
}
func (b backend) RemoveRange(r NumberRange) error {
	return nil
}
func (b backend) AddRange(r NumberRange) ([]NumberRange, error) {
	return nil, nil
}
func (b backend) RangeFor(number uint64) (NumberRange, error) {
	if number == 4741067196 {
		return NumberRange{Upper: 4741067196, Lower: 4741067196, Regexp: "!^(.*)$!sip:\\1@closed.sip..net!"}, nil
	} else {
		return NumberRange{Upper: number, Lower: number, Regexp: "!^(.*)$!sip:\\1@v9.sip..net!"}, nil
	}
}
