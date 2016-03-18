# enum-dns

[![Build Status](https://travis-ci.org/hadrienk/enum-dns.svg?branch=master)](https://travis-ci.org/hadrienk/enum-dns)

Enum-dns is a simple ENUM DNS server written in Go that is made to handle ENUM requests only.
 
Many DNS server implementations support NAPTR and thus can serve as an ENUM server. However, they usually store records in a flat database. In the case of ENUM, this can get highly unpractical since a common use case is to route SIP requests directed to a range of numbers to a location.

Enum-dns stores the records together with "intervals" (or ranges). For every ENUM query it receives, enum-dns will try to find an interval the query falls into and return the corresponding record.
  
It relies on a simple interface to query the data. You can make it fit your current system by implementing the following: 
  
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

Enum-dns even comes with default backend implementations. If you decide to make one that fits your needs, free to make a pull request.
 
* MySQL
* Memory backend

## Interval model

TODO  

## Caching and security

If you don't want to expose enum-dns directly (and you sould not!), you can easily set up another DNS server like named, dnsmasq or unbound in front of it.
 
 Running a DNS server in front of enum-dns allows you to secure your server and improve performance using it as a cache for instance.

The following configuration examples require enum dns to run on the same machine and listen on the address 127.0.0.1 and port 5354.

### Unbound

Unbound is one of the few DNS cache server that supports forcing minimum TTL. Forcing minimum cache is not recommended outside of particular scenarios like this one. 

```
/etc/unbound/unbound.conf

server:
        verbosity: 1
        interface: 0.0.0.0
        do-ip4: yes
        do-udp: yes
        
        hide-identity: yes
        hide-version: yes
        
        access-control: 0.0.0.0/0 allow
        
        cache-max-ttl: 3600 # 60 min maximum cache
        cache-min-ttl: 900  # 15 min minimum cache
        
        # The local IP address is is blacklisted by default (loop prevention).
        # In our case, we want to be able to query 127.0.0.1
        do-not-query-localhost: no
        
forward-zone:
        name: "e164.arpa."
        forward-addr: 127.0.0.1@5354
        
```

### named

```
// /etc/bind/named.conf

options {
    directory "/var/cache/bind";
    
    allow-query-cache { any; };
    
    // This option is important to avoid additional information
    // to be added to the response.
    minimal-responses yes;
    
};

zone "e164.arpa" {
    type forward;
    forward only;
    forwarders { 127.0.0.1 port 5354; };
};

```

### dnsmasq

```
# /etc/dnsmasq.conf

# ...

server=/e164.arpa/127.0.0.1#5354

# ...
```

## Configuration

TODO