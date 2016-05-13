# Go gRPC Interceptors for Prometheus monitoring 

[![Travis Build](https://travis-ci.org/mwitkow/go-flagz.svg)](https://travis-ci.org/mwitkow/go-grpc-prometheus)
[![Go Report Card](http://goreportcard.com/badge/mwitkow/go-flagz)](http://goreportcard.com/report/mwitkow/go-grpc-prometheus)
[![GoDoc](http://img.shields.io/badge/GoDoc-Reference-blue.svg)](https://godoc.org/github.com/mwitkow/go-grpc-prometheus)
[![Apache 2.0 License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

[Prometheus](https://prometheus.io/) monitoring for your [gRPC Go](https://github.com/grpc/grpc-go) servers.

## Interceptors

[gRPC Go](https://github.com/grpc/grpc-go) recently acquired support for Interceptors, i.e. middleware that is executed
by a gRPC Server before the request is passed onto the user's application logic. It is a perfect way to implement
common patters: auth, logging and... monitoring.

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

## Status

This code has been in an upstream [pull request](https://github.com/grpc/grpc-go/pull/299) since August 2015. It has 
served as the basis for monitoring of production gRPC micro services at [Improbable](https://improbable.io) since then.

## License

`go-grpc-prometheus` is released under the Apache 2.0 license. See the [LICENSE](LICENSE) file for details.
