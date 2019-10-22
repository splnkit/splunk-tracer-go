package splunktracing_test

import (
	"context"

	"github.com/lightstep/lightstep-tracer-common/golang/gogo/collectorpb"
	"github.com/lightstep/lightstep-tracer-common/golang/gogo/collectorpb/collectorpbfakes"
	"github.com/lightstep/lightstep-tracer-go"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/opentracing/opentracing-go"
)

var _ = Describe("Tracerv0_14", func() {
	var tracer lightstep.Tracerv0_14
	var opts lightstep.Options

	const accessToken = "ACCESS_TOKEN"

	var fakeClient *collectorpbfakes.FakeCollectorServiceClient
	var fakeConn lightstep.ConnectorFactory

	var eventHandler lightstep.EventHandler
	const eventBufferSize = 10

	BeforeEach(func() {
		fakeClient = new(collectorpbfakes.FakeCollectorServiceClient)
		fakeClient.ReportReturns(&collectorpb.ReportResponse{}, nil)
		fakeConn = fakeGrpcConnection(fakeClient)

		eventHandler, _ = lightstep.NewEventChannel(eventBufferSize)
		lightstep.SetGlobalEventHandler(eventHandler)
	})

	JustBeforeEach(func() {
		opts = lightstep.Options{
			AccessToken: accessToken,
			ConnFactory: fakeConn,
		}
		tracer = lightstep.NewTracerv0_14(opts)
	})

	AfterEach(func() {
		closeTestTracer(tracer)
	})

	Describe("Helper functions", func() {
		It("GetSplunkAccessToken returns the access token", func() {
			Expect(lightstep.GetSplunkAccessToken(tracer)).To(Equal(accessToken))
		})

		It("Close closes the tracer", func() {
			lightstep.Close(context.Background(), tracer)
		})

		It("CloseTracer closes the tracer", func() {
			err := lightstep.CloseTracer(tracer)
			Expect(err).ToNot(HaveOccurred())
		})

		It("Flush flushes the tracer", func() {
			lightstep.Flush(context.Background(), tracer)
		})

		It("FlushSplunkTracer flushes the tracer", func() {
			err := lightstep.FlushSplunkTracer(tracer)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("provides its ReporterID", func() {
		It("is non-zero", func() {
			rid, err := lightstep.GetSplunkReporterID(opentracing.Tracer(tracer))
			Expect(err).To(BeNil())
			Expect(rid).To(Not(BeZero()))
		})
	})
})
