# graphite-client 
[![Build Status](https://travis-ci.org/gguridi/graphite-client.svg?branch=master)](https://travis-ci.org/gguridi/graphite-client)
[![GoDoc](https://godoc.org/github.com/gguridi/graphite-client?status.svg)](https://godoc.org/github.com/gguridi/graphite-client)

Graphite client written in Go, focused in thread-safe and automatic aggregations.

## Installation

Use `go get` or your favourite dependency manager to install it from github:

```bash
go get github.com/gguridi/graphite-client
```

## Dependencies

This library doesn't have external dependencies by itself, only for testing purposes
we need [Gingko](https://github.com/onsi/ginkgo) and [Gomega](https://github.com/onsi/gomega).

## Configuration

The different options to configure the client can be found [here]().

## Protocols

Currently we support two protocols to connect with graphite.

- **TCP**: using the `NewGraphiteTCP` constructor.
- **UDP**: using the `NewGraphiteUDP` constructor.

## Simple client

We can initialise a simple client with one of the constructors:

```go
import (
    graphite "github.com/gguridi/graphite-client"
    "fmt"
)

client := graphite.NewGraphiteTCP(graphite.Config{
    Host: "example.com",
    Port: 2003,
})
```

and then send one metric to graphite:

```go
if _, err:= client.Send("metric.name.count", 55); err != nil {
    fmt.Println("Unable to send metrics to graphite")
}
```

or if we don't mind if we could send or not the metric to graphite:

```go
client.Send("metric.name.count", 55)
```

## Aggregator

We can use an aggregator to send more than one metric at a time, for systems that collect
a lot of data and we don't want to overload graphite with requests.

We can initialise an aggregator from any of the supported clients:

```go
import (
    graphite "github.com/gguridi/graphite-client"
    "fmt"
)

aggregator := graphite.NewGraphiteTCP(graphite.Config{
    Host: "example.com",
    Port: 2003,
}).NewAggregator()
```

and then send several metric to graphite:

```go
aggregator.AddSum("metric.received.count", 15)
aggregator.AddSum("metric.received.count", 10)
aggregator.AddSum("metric.sent.count", 5)
aggregator.Flush()
```

This will send two metrics at once, one for `metric.received.count` with a value of 25, and
another one for `metric.sent.count` with a value of 5.

### Metric types

There are several kind of metrics built-in in the system, that can be automatically used
with the method of the `Aggregator` interface.

- `AddSum`: Will initialise a metric where the final value sent to graphite will be the addition
of all the values passed to the aggregator. So if we call `AddSum` with a specific metric path and
values 5, 10, 15 and then we `Flush`, we will be sending a final value of 30 to graphite.
- `Increase`: Used as an alias of `AddSum` where the value incremented is always 1. Useful for giving
a comprehensive behaviour to the metric. 
- `AddAverage`: Will initialise a metric where the final value sent to graphite will be the average 
of all the values passed to the aggregator. So if we call `AddAverage` with a specific metric path
and values 2, 10, 10 and then we `Flush`, we will be sending a final value of 7.333333 to graphite. The
maximum decimals allowed is 6.
- `SetActive`/`SetInactive`: Will initialise a metric where the final value sent to graphite will 
be 1 (`SetActive`) or 0 (`SetInactive`). This way we can send metrics such service status, etc.

### Automatic flush

It's possible to configure the aggregator to periodically flush the values to graphite without
having to worry about doing it manually. To make it possible we can initialise the aggregator with:

```go
aggregator := graphite.NewGraphiteTCP(graphite.Config{
    Host: "example.com",
    Port: 2003,
}).NewAggregator().Run(30 * time.Second, nil)
```

The second parameter that method accepts is a `chan bool` that we can use to stop the loop whenever we want.

```go
stop := make(chan bool)

aggregator := graphite.NewGraphiteTCP(graphite.Config{
    Host: "example.com",
    Port: 2003,
}).NewAggregator().Run(30 * time.Second, stop)

time.Sleep(600)
stop <- true
```
