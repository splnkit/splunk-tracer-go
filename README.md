# splunk-tracer-go

[![Circle CI](https://circleci.com/gh/splunk/splunk-tracer-go.svg?style=shield)](https://circleci.com/gh/splunk/splunk-tracer-go)
[![MIT license](http://img.shields.io/badge/license-MIT-blue.svg)](http://opensource.org/licenses/MIT)
[![GoDoc](https://godoc.org/github.com/splunk/splunk-tracer-go?status.svg)](https://godoc.org/github.com/splunk/splunk-tracer-go)

The splunk distributed tracing library for Go.

**Looking for the splunk OpenCensus exporter? Check out the [`splunkoc` package](./splunkoc).**

## Installation

```
$ go get 'github.com/splunk/splunk-tracer-go'
```

## API Documentation

Godoc: https://godoc.org/github.com/splunk/splunk-tracer-go

## Initialization: Starting a new tracer
To initialize a tracer, configure it with a valid Access Token and optional tuning parameters. Register the tracer as the OpenTracing global tracer so that it will become available to your installed instrumentation libraries.

```go
import (
  "github.com/opentracing/opentracing-go"
  "github.com/splunk/splunk-tracer-go"
)

func main() {
  splunkTracer := splunktracing.NewTracer(splunk.Options{
    AccessToken: "YourAccessToken",
  })

  opentracing.SetGlobalTracer(splunkTracer)
}
```

## Instrumenting Code: Using the OpenTracing API

All instrumentation should be done through the OpenTracing API, rather than using the splunk tracer type directly. For API documentation and advice on instrumentation in general, see the opentracing godocs and the opentracing website.

- https://godoc.org/github.com/opentracing/opentracing-go
- http://opentracing.io

## Flushing and Closing: Managing the tracer lifecycle

As part of managing your application lifecycle, the splunk tracer extends the `opentracing.Tracer` interface with methods for manual flushing and closing. To access these methods, you can take the global tracer and typecast it to a `splunk.Tracer`. As a convenience, the splunk package provides static methods which perform the typecasting.

```go
import (
  "context"
  "github.com/opentracing/opentracing-go"
  "github.com/splunk/splunk-tracer-go"
)

func shutdown(ctx context.Context) {
  // access the running tracer
  tracer := opentracing.GlobalTracer()
    
  // typecast from opentracing.Tracer to splunk.Tracer
  splTracer, ok := tracer.(splunktracing.Tracer)
  if (!ok) { 
    return 
  }
  splTracer.Close(ctx)

  // or use static methods
  splunktracing.Close(ctx, tracer)
}
```

## Event Handling: Observing the splunk tracer
In order to connect diagnostic information from the splunk tracer into an application's logging and metrics systems, inject an event handler using the `OnEvent` static method. Events may be typecast to check for errors or specific events such as status reports.

```go
import (
  "example/logger"
  "example/metrics"
  "github.com/splunk/splunk-tracer-go"
)

logAndMetricsHandler := func(event splunktracing.Event){
  switch event := event.(type) {
  case EventStatusReport:
    metrics.Count("tracer.dropped_spans", event.DroppedSpans())
  case ErrorEvent:
    logger.Error("Splunk Tracer error: %s", event)
  default:
    logger.Info("Splunk Tracer info: %s", event)
  }
}

func main() {
  // setup event handler first to catch startup errors
  splunktracing.SetGlobalEventHandler(logAndMetricsHandler)
  
  splunkTracer := splunktracing.NewTracer(splunk.Options{
    AccessToken: "YourAccessToken",
  })

  opentracing.SetGlobalTracer(splunkTracer)
}
```

Event handlers will receive events from any active tracers, as well as errors in static functions. It is suggested that you set up event handling before initializing your tracer to catch any errors on initialization.
