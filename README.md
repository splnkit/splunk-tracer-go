# splunk-tracer-go


[![MIT license](http://img.shields.io/badge/license-MIT-blue.svg)](http://opensource.org/licenses/MIT)


The Splunk distributed tracing library for Go.


## Installation

```
$ go get 'github.com/splnkit/splunk-tracer-go'
```

## API Documentation


## Initialization: Starting a new tracer
To initialize a tracer, configure it with a valid Access Token and optional tuning parameters. Register the tracer as the OpenTracing global tracer so that it will become available to your installed instrumentation libraries.

```go
import (
  "github.com/opentracing/opentracing-go"
  "github.com/splnkit/splunk-tracer-go"
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

As part of managing your application lifecycle, the Splunk tracer extends the `opentracing.Tracer` interface with methods for manual flushing and closing. To access these methods, you can take the global tracer and typecast it to a `splunktracing.Tracer`. As a convenience, the splunk package provides static methods which perform the typecasting.

```go
import (
  "context"
  "github.com/opentracing/opentracing-go"
  "github.com/splnkit/splunk-tracer-go"
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



This library is the Splunk binding for [OpenTracing](http://opentracing.io/). See the [OpenTracing Go API](https://github.com/opentracing/opentracing-go) for additional detail.

## License

The Splunk Tracer for Go is licensed under the MIT License. Details can be found in the LICENSE file.

### Third-party libraries

This is a fork of the Go tracer from Lightstep, which is also licensed under the MIT License. Links to the original repository and license are below:

* [lightstep-tracer-go][lightstep]: [MIT][lightstep-license]

[lightstep]:                      https://github.com/lightstep/lightstep-tracer-go
[lightstep-license]:              https://github.com/lightstep/lightstep-tracer-go/blob/master/LICENSE