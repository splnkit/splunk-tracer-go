// Based on the quickstart tutorial available at
// https://opencensus.io/quickstart/go/tracing
package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/splunk/splunk-tracer-go/splunkoc"
	"go.opencensus.io/trace"
)

const (
	DefaultSatelliteHost = "localhost"
	DefaultSatellitePort = 8360
)

var (
	accessToken   = flag.String("access-token", "", "Splunk access token")
	host          = flag.String("host", "", "Splunk HTTP Event Collector host")
	port          = flag.Int("port", 0, "Splunk HTTP Event Collector port")
	insecure      = flag.Bool("insecure", false, "Use an insecure connection when connection to the Splunk HEC endpoint")
	componentName = flag.String("component-name", "", "Component name")
)

func main() {
	flag.Parse()

	if *host == "" {
		*host = DefaultSatelliteHost
	}
	if *port == 0 {
		*port = DefaultSatellitePort
	}

	exporterOptions := []splunkoc.Option{
		splunkoc.WithAccessToken(*accessToken),
		splunkoc.WithSatelliteHost(*host),
		splunkoc.WithSatellitePort(*port),
		splunkoc.WithInsecure(*insecure),
		splunkoc.WithComponentName(*componentName),
	}

	exporter, err := splunkoc.NewExporter(exporterOptions...)
	if err != nil {
		log.Fatal(err)
	}
	defer exporter.Close(context.Background())

	trace.RegisterExporter(exporter)

	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})

	ctx, span := trace.StartSpan(context.Background(), "main")
	defer span.End()

	for i := 0; i < 10; i++ {
		doWork(ctx)
	}

	exporter.Flush(context.Background())
}

func doWork(ctx context.Context) {
	_, span := trace.StartSpan(ctx, "doWork")
	defer span.End()

	fmt.Println("doing busy work")
	time.Sleep(80 * time.Millisecond)
	buf := bytes.NewBuffer([]byte{0xFF, 0x00, 0x00, 0x00})
	num, err := binary.ReadVarint(buf)
	if err != nil {
		span.SetStatus(trace.Status{
			Code:    trace.StatusCodeUnknown,
			Message: err.Error(),
		})
	}

	span.Annotate([]trace.Attribute{
		trace.Int64Attribute("bytes to int", num),
	}, "Invoking doWork")
	time.Sleep(20 * time.Millisecond)
}
