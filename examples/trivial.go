// A trivial Splunk Go tracer example.
//
// $ go build -o splunk_trivial github.com/splunk/splunk-tracer-go/examples/trivial
// $ ./splunk_trivial --access_token=YOUR_ACCESS_TOKEN

package main

import (
	"context"
	"flag"
	"fmt"
	logger "log"
	"time"

	"github.com/splnkit/splunk-tracer-go"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
)

func subRoutine(ctx context.Context) {
	trivialSpan, _ := opentracing.StartSpanFromContext(ctx, "test span")
	defer trivialSpan.Finish()
	trivialSpan.LogEvent("logged something")
	trivialSpan.LogFields(log.String("string_key", "some string value"), log.Object("trivialSpan", trivialSpan))

	subSpan := opentracing.StartSpan(
		"child span", opentracing.ChildOf(trivialSpan.Context()))
	trivialSpan.LogFields(log.Int("int_key", 42), log.Object("subSpan", subSpan),
		log.String("time.eager", fmt.Sprint(time.Now())),
		log.Lazy(func(fv log.Encoder) {
			fv.EmitString("time.lazy", fmt.Sprint(time.Now()))
		}))
	defer subSpan.Finish()
}

type LoggingRecorder struct {
	r splunktracing.SpanRecorder
}

func (r *LoggingRecorder) RecordSpan(span splunktracing.RawSpan) {
	logger.Printf("span traceID: %v spanID: %v parentID: %v Operation: %v \n", span.Context.TraceID, span.Context.SpanID, span.ParentSpanID, span.Operation)
}

func main() {
	accessToken := flag.String("access_token", "", "your Splunk access token")
	flag.Parse()
	loggableRecorder := &LoggingRecorder{}

	// Use Splunk as the global OpenTracing Tracer.
	opentracing.InitGlobalTracer(splunktracing.NewTracer(splunktracing.Options{
		AccessToken: *accessToken,
		Collector:   splunktracing.Endpoint{Host: "127.0.0.1", Port: 8088, Plaintext: false},
		Recorder:    loggableRecorder,
	}))

	fmt.Println(*accessToken)

	// Do something that's traced.
	subRoutine(context.Background())

	// Force a flush before exit.
	splunktracing.Flush(context.Background(), opentracing.GlobalTracer())
}
