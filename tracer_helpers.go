package splunktracing

import (
	"context"

	"github.com/opentracing/opentracing-go"
)

// Flush forces a synchronous Flush.
func Flush(ctx context.Context, tracer opentracing.Tracer) {
	switch splkTracer := tracer.(type) {
	case Tracer:
		splkTracer.Flush(ctx)
	case *tracerv0_14:
		Flush(ctx, splkTracer.Tracer)
	default:
		emitEvent(newEventUnsupportedTracer(tracer))
	}
}

// Close synchronously flushes the tracer, then terminates it.
func Close(ctx context.Context, tracer opentracing.Tracer) {
	switch splkTracer := tracer.(type) {
	case Tracer:
		splkTracer.Close(ctx)
	case *tracerv0_14:
		Close(ctx, splkTracer.Tracer)
	default:
		emitEvent(newEventUnsupportedTracer(tracer))
	}
}

// GetSplunkAccessToken returns the currently configured AccessToken.
func GetSplunkAccessToken(tracer opentracing.Tracer) (string, error) {
	switch splkTracer := tracer.(type) {
	case Tracer:
		return splkTracer.Options().AccessToken, nil
	case *tracerv0_14:
		return GetSplunkAccessToken(splkTracer.Tracer)
	default:
		return "", newEventUnsupportedTracer(tracer)
	}
}

// FlushSplunkTracer flushes the tracer
// DEPRECATED: use Flush instead.
func FlushSplunkTracer(tracer opentracing.Tracer) error {
	switch splkTracer := tracer.(type) {
	case Tracer:
		splkTracer.Flush(context.Background())
		return nil
	case *tracerv0_14:
		return FlushSplunkTracer(splkTracer.Tracer)
	default:
		return newEventUnsupportedTracer(tracer)
	}
}

// CloseTracer closes the tracer
// DEPRECATED: use Close instead.
func CloseTracer(tracer opentracing.Tracer) error {
	switch splkTracer := tracer.(type) {
	case Tracer:
		splkTracer.Close(context.Background())
		return nil
	case *tracerv0_14:
		return CloseTracer(splkTracer.Tracer)
	default:
		return newEventUnsupportedTracer(tracer)
	}
}

// GetSplunkReporterID returns the currently configured Reporter ID.
func GetSplunkReporterID(tracer opentracing.Tracer) (uint64, error) {
	switch splkTracer := tracer.(type) {
	case *tracerImpl:
		return splkTracer.reporterID, nil
	case *tracerv0_14:
		return GetSplunkReporterID(splkTracer.Tracer)
	default:
		return 0, newEventUnsupportedTracer(tracer)
	}
}
