# enum-dns

Enum-dns is a simple enum dns server written in Go meant to handle ENUM DNS requests only.
 
Many DNS servers support NAPTR and thus ENUM, however, they usually store every record. In the case of ENUM, this can get highly unpractical.

Enum-dns stores the records "intervals" (or ranges). For every enum query it receives, enum-dns will try to find an interval the query falls into.  

  
```go
package enum

type Backend interface {
	
	Ranges(n uint64, c int) ([]NumberRange, error)

	RangeFor(n uint64) (NumberRange, error)

	AddRange(r NumberRange) ([]NumberRange, error)

	RemoveRange(r NumberRange) error

	Close() error
}
```
