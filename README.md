# enum-dns

[![Build Status](https://travis-ci.org/hadrienk/enum-dns.svg?branch=master)](https://travis-ci.org/hadrienk/enum-dns)

Enum-dns is a simple ENUM DNS server written in Go that is made to handle ENUM requests only.
 
Many DNS server implementations support NAPTR and thus can serve as an ENUM server. However, they usually store records in a flat database. In the case of ENUM, this can get highly unpractical since a common use case is to route SIP requests directed to a range of numbers to a location.

Enum-dns stores the records together with "intervals" (or ranges). For every ENUM query it receives, enum-dns will try to find an interval the query falls into and return the corresponding record.
  
It relies on a simple interface to query the data. You can make it fit your current system by implementing the following: 
  
```go
type Backend interface {
    	// RangesBetween returns a list of ranges that enclose the given range l(ower) to u(pper) or
	// nil if no range matches.
	// The c parameter is the maximum count of values to return. If a negative c value is used
	// it will return the ranges in reverse order.
	RangesBetween(l, u uint64, c int) ([]NumberRange, error)

	// Add a range to the backend. Any range overlapping with the one added will be deleted or
	// adjusted to make room for the new one and returned.
	PushRange(r NumberRange) ([]NumberRange, error)

	// Close the backend.
	Close() error
}
```

## Rest API

Enum-dns also comes with a REST API to manipulate the backend's data. 

### `/inverval/{from}:{to}`

#### Parameters

 *required*
 from: number between 10000000000000000 and 9999999999999999
  
  *required*
  to: number between 10000000000000000 and 9999999999999999 

#### Methods
  
  GET: Return the interval and its record. Return 404 if no interval for the passed from and to parameters exist.
  
  PUT: Create a new interval. Returns 201 if creation succeeded, and an array of the intervals that were overwritten.
  
  Content: 
  
```json
  {
          "upper":100000858306882,
          "lower":100000000000000,
          "records":[
             {
                "order":10,
                "preference":100,
                "flags":"",
                "service":"E2U+sip",
                "regexp":"!^(.*)$!sip:\\@default!",
                "replacement":"."
             }
          ]
       }
```  
  Example content
  
```json
  [
     {
        "upper":100000858306882,
        "lower":100000000000000,
        "records":[
           {
              "order":10,
              "preference":100,
              "flags":"",
              "service":"E2U+sip",
              "regexp":"!^(.*)$!sip:\\@default!",
              "replacement":"."
           }
        ]
     }
  ]
  
```

 
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

### Named (Bind 9)

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

### Dnsmasq

```
# /etc/dnsmasq.conf

# ...

server=/e164.arpa/127.0.0.1#5354

# ...
```

## Configuration

TODO
