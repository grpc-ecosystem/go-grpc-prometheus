# Go gRPC Interceptors for Prometheus monitoring 

[![Travis Build](https://travis-ci.org/mwitkow/go-grpc-prometheus.svg)](https://travis-ci.org/mwitkow/go-grpc-prometheus)
[![Go Report Card](http://goreportcard.com/badge/mwitkow/go-grpc-prometheus)](http://goreportcard.com/report/mwitkow/go-grpc-prometheus)
[![GoDoc](http://img.shields.io/badge/GoDoc-Reference-blue.svg)](https://godoc.org/github.com/mwitkow/go-grpc-prometheus)
[![Apache 2.0 License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

[Prometheus](https://prometheus.io/) monitoring for your [gRPC Go](https://github.com/grpc/grpc-go) servers.

## Interceptors

[gRPC Go](https://github.com/grpc/grpc-go) recently acquired support for Interceptors, i.e. middleware that is executed
by a gRPC Server before the request is passed onto the user's application logic. It is a perfect way to implement
common patterns: auth, logging and... monitoring.

To use Interceptors in chains, please see (TODO: publish chaining interceptors lib).

## Usage

```go
import "github.com/mwitkow/go-grpc-prometheus"
...

    myServer := grpc.NewServer(
        grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
        grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
    )
...
```

# Metrics


## Labels

All server-side metrics start with `grpc_server` as Prometheus subsystem name. Similarly all methods
contain the same rich labels:
  
  * `service` - the [gRPC service](http://www.grpc.io/docs/#defining-a-service) name, which is the combination of protobuf `package` and
    the `service` section name. E.g. for `package = mwitkow.testproto` and 
     `service TestService` the label will be `service="mwitkow.testproto.TestService"`
  * `method` - the name of the method called on the gRPC service. E.g.  
    `method="Ping"`
  * `type` - the gRPC [type of request](http://www.grpc.io/docs/guides/concepts.html#rpc-life-cycle). 
    Differentiating between the two is important especially for latency measurements.

     - `unary` is single request, single response RPC
     - `client_stream` is a multi-request, single response RPC
     - `server_stream` is a single request, multi-response RPC
     - `bidi_stream` is a multi-request, multi-response RPC
    

Additionally for completed RPCs, the following labels are used:

  * `code` - the human-readable [gRPC status code](https://github.com/grpc/grpc-go/blob/master/codes/codes.go).
    The list of all statuses is to long, but here are some common ones:
      
      - `OK` - means the RPC was successful
      - `IllegalArgument` - RPC contained bad values
      - `Internal` - server-side error not disclosed to the clients
      
## Counters

The counters and their up to date documentation is in [server_reporter.go](server_reporter.go) and
the respective Prometheus handler (usually `/metrics`). 

For simplicity, let's assume we're tracking a single server-side RPC call of [`mwitkow.testproto.TestService`](examples/testproto/test.proto),
calling the method `PingList`. The call succeeds and returns 20 messages in the stream.

First, immediately after the server receives the call it will increment the
`grpc_server_rpc_started_total` and start the handling time clock. 

```jsoniq
grpc_server_rpc_started_total{method="PingList",service="mwitkow.testproto.TestService",type="server_stream"} 1
```

Then the user logic gets invoked. It receives one message from the client containing the request 
(it's a `server_stream`):

```jsoniq
grpc_server_rpc_msg_received_total{method="PingList",service="mwitkow.testproto.TestService",type="server_stream"} 1
```

The user logic may return an error, or send multiple messages back to the client. In this case, on 
each of the 20 messages sent back, a counter will be incremented:

```jsoniq
grpc_server_rpc_msg_sent_total{method="PingList",service="mwitkow.testproto.TestService",type="server_stream"} 20
```

After the call completes, it's status (`OK` or error code) gets represented in the [Prometheus histogram](https://prometheus.io/docs/concepts/metric_types/#histogram)
variable `grpc_server_rpc_handled`. It contains three sub-metrics:

 * `grpc_server_rpc_handled_count` - the count of all completed RPCs by status and method 
 * `grpc_server_rpc_handled_sum` - cumulative time of RPCs by status and method, useful for 
   calculating average handling times
 * `grpc_server_rpc_handled_bucket` - contains the counts of RPCs by status and method in respective
   handling-time buckets. These buckets can be used by Prometheus to estimate SLAs (see [here](https://prometheus.io/docs/practices/histograms/))

The counter values will look as follows:

```jsoniq
grpc_server_rpc_handled_bucket{code="OK",method="PingList",service="mwitkow.testproto.TestService",type="server_stream",le="0.005"} 1
grpc_server_rpc_handled_bucket{code="OK",method="PingList",service="mwitkow.testproto.TestService",type="server_stream",le="0.01"} 1
grpc_server_rpc_handled_bucket{code="OK",method="PingList",service="mwitkow.testproto.TestService",type="server_stream",le="0.025"} 1
grpc_server_rpc_handled_bucket{code="OK",method="PingList",service="mwitkow.testproto.TestService",type="server_stream",le="0.05"} 1
grpc_server_rpc_handled_bucket{code="OK",method="PingList",service="mwitkow.testproto.TestService",type="server_stream",le="0.1"} 1
grpc_server_rpc_handled_bucket{code="OK",method="PingList",service="mwitkow.testproto.TestService",type="server_stream",le="0.25"} 1
grpc_server_rpc_handled_bucket{code="OK",method="PingList",service="mwitkow.testproto.TestService",type="server_stream",le="0.5"} 1
grpc_server_rpc_handled_bucket{code="OK",method="PingList",service="mwitkow.testproto.TestService",type="server_stream",le="1"} 1
grpc_server_rpc_handled_bucket{code="OK",method="PingList",service="mwitkow.testproto.TestService",type="server_stream",le="2.5"} 1
grpc_server_rpc_handled_bucket{code="OK",method="PingList",service="mwitkow.testproto.TestService",type="server_stream",le="5"} 1
grpc_server_rpc_handled_bucket{code="OK",method="PingList",service="mwitkow.testproto.TestService",type="server_stream",le="10"} 1
grpc_server_rpc_handled_bucket{code="OK",method="PingList",service="mwitkow.testproto.TestService",type="server_stream",le="+Inf"} 1
grpc_server_rpc_handled_sum{code="OK",method="PingList",service="mwitkow.testproto.TestService",type="server_stream"} 0.0003866430000000001
grpc_server_rpc_handled_count{code="OK",method="PingList",service="mwitkow.testproto.TestService",type="server_stream"} 1
```


## Useful query examples

Prometheus philosophy is to provide the most detailed metrics possible to the monitoring system, and
let the aggregations be handled there. The verbosity of above metrics make it possible to have that
flexibility. Here's a couple of useful monitoring queries:


### request inbound rate
```jsoniq
sum(rate(grpc_server_rpc_started_total{job="foo"}[1m])) by (service)
```
For `job="foo"` (common label to differentiate between Prometheus monitoring targets), calculate the
rate of requests per second (1 minute window) for each gRPC `service` that the job has. Please note
how the `method` is being omitted here: all methods of a given gRPC service will be summed together.

### unary request error rate
```jsoniq
sum(rate(grpc_server_rpc_handled_count{job="foo",type="unary",code!="OK"}[1m])) by (service)
```
For `job="foo"`, calculate the per-`service` rate of `unary` (1:1) RPCs that failed, i.e. the 
ones that didn't finish with `OK` code.

### unary request error percentage
```jsoniq
sum(rate(grpc_server_rpc_handled_count{job="foo",type="unary",code!="OK"}[1m])) by (service)
 / 
sum(rate(grpc_server_rpc_started_total{job="foo",type="unary"}[1m])) by (service)
 * 100.0
```
For `job="foo"`, calculate the percentage of failed requests by service. It's easy to notice that
this is a combination of the two above examples. This is an example of a query you would like to
[alert on](https://prometheus.io/docs/alerting/rules/) in your system for SLA violations, e.g.
"no more than 1% requests should fail".

### average response stream size
```jsoniq
sum(rate(grpc_server_rpc_msg_sent_total{job="foo",type="server_stream"}[10m])) by (service)
 /
sum(rate(grpc_server_rpc_handled_count{job="foo",type="server_stream",code="OK"}[10m])) by (service)
```
For `job="foo"` what is the `service`-wide `10m` average of messages returned for all `server_stream` 
RPCs. This allows you to track the stream sizes returned by your system, e.g. allows you
to track when clients started to send "wide" queries that return hundreds of responses.

### 99%-tile latency of unary requests
```jsoniq
histogram_quantile(0.99, 
  sum(rate(grpc_server_rpc_handled_bucket{job="foo",type="unary"}[5m])) by (service,le)
)
```
For `job="foo"`, returns an 99%-tile [quantile estimation](https://prometheus.io/docs/practices/histograms/#quantiles)
of the handling time of RPCs per service. Please note the `5m` rate, this means that the quantile
estimation will take samples in a rolling `5m` window. When combined with other quantiles
(e.g. 50%, 90%), this query gives you tremendous insight into the responsiveness of your system 
(e.g. impact of caching).

### percentage of slow unary queries (>250ms)
```jsoniq
100.0 - (
sum(rate(grpc_server_rpc_handled_bucket{job="foo",type="unary",le="0.25"}[5m])) by (service)
 / 
sum(rate(grpc_server_rpc_handled_count{job="foo",type="unary"}[5m])) by (service)
) * 100.0
```
For `job="foo"` calculate the by-`service` fraction of slow requests that took longer than `0.25` 
seconds. This query is relatively complex, since the Prometheus aggregations use `le` (less or equal)
buckets, meaning that counting "fast" requests fractions is easier. However, simple maths helps.
This is an example of a query you would like to alert on in your system for SLA violations, 
e.g. "less than 1% of requests are slower than 250ms".


## Status

This code has been in an upstream [pull request](https://github.com/grpc/grpc-go/pull/299) since August 2015. It has 
served as the basis for monitoring of production gRPC micro services at [Improbable](https://improbable.io) since then.

## License

`go-grpc-prometheus` is released under the Apache 2.0 license. See the [LICENSE](LICENSE) file for details.
