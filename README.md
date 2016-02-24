# enum-dns

Enum-dns is a simple ENUM DNS server written in Go that is made to handle ENUM requests only.
 
Many DNS server implementations support NAPTR and thus can serve as an ENUM server. However, they usually store records in a flat database. In the case of ENUM, this can get highly unpractical since a common use case scenario is to route SIP requests directed to a range of numbers to a location.

Enum-dns stores the records together with "intervals" (or ranges). For every ENUM query it receives, enum-dns will try to find an interval the query falls into and return the corresponding record.
  
It relies on a single simple interface to query the data. Integrating it so that it can fit you current system is a matter of implementing the following: 
  
```go

type Backend interface {
	
	Ranges(n uint64, c int) ([]NumberRange, error)

	RangeFor(n uint64) (NumberRange, error)

	AddRange(r NumberRange) ([]NumberRange, error)

	RemoveRange(r NumberRange) error

	Close() error
}
```

## Existing backends

Enum-dns even comes with default backends. If you don't find what you need and make one, free to make a pull request.
 
* MySQL
* BoltDB
* Static

## Interval model

TODO  

## Caching

If you don't want to expose enum-dns directly, you can easily set up another DNS server like named or dnsmasq in front of it.  

### named

### dnsmasq

## Configuration

TODO