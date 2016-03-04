# enum-dns

Enum-dns is a simple ENUM DNS server written in Go that is made to handle ENUM requests only.
 
Many DNS server implementations support NAPTR and thus can serve as an ENUM server. However, they usually store records in a flat database. In the case of ENUM, this can get highly unpractical since a common use case scenario is to route SIP requests directed to a range of numbers to a location.

Enum-dns stores the records together with "intervals" (or ranges). For every ENUM query it receives, enum-dns will try to find an interval the query falls into and return the corresponding record.
  
It relies on a single simple interface to query the data. Integrating it so that it can fit you current system is a matter of implementing the following: 
  
```go
type Backend interface {

	RangesBetween(l, u uint64, c int) ([]NumberRange, error)

	AddRange(r NumberRange) ([]NumberRange, error)

	RemoveRange(r NumberRange) error

	Close() error
}
```

## Rest API

Enum-dns also comes with a REST API to manipulate the backend's data. 

/inverval GET

`/inverval/{prefix},[{limit}]` GET|UPDATE

prefix: String [1-9][0-9]{0,13}
limit: String -?[0-9]+

Example: 

Calling '/interval/47,10' will return the 10 *first* intervals that overlap with [470000000000000,479999999999999] in ascending order.

Calling '/interval/474,-10' will return the 10 *last* intervals that overlap with [474000000000000,474999999999999] in descending order.

/inverval/{from}[,{to}[,{limit}]] GET|UPDATE

from, to: String [1-9][0-9]{0,14} 
limit: String -?[0-9]+

Example:

Calling /inverval/471234567800000 will return all the intervals that match [471234567800000,471234567800000]. One in that case.

Calling /inverval/471234567800000,472000000000000 will return all the intervals that match [471234567800000,472000000000000] in ascending order.

Calling /inverval/471234567800000,472000000000000,- will return all the intervals that match [471234567800000,472000000000000] in descending order.
 
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