package splunktracing

import (
	"bytes"
	"encoding/json"
	// "fmt"
	"strconv"
	"strings"
	"time"

	// "github.com/opentracing/opentracing-go"
)

type hecConverter struct {
	verbose        bool
	maxLogKeyLen   int // see GrpcOptions.MaxLogKeyLen
	maxLogValueLen int // see GrpcOptions.MaxLogValueLen
}

type splLog struct {
	timestamp          time.Time
	span_id            string
	trace_id		   string
}

func newHECConverter(options Options) *hecConverter {
	return &hecConverter{
		verbose:        options.Verbose,
		maxLogKeyLen:   options.MaxLogKeyLen,
		maxLogValueLen: options.MaxLogValueLen,
	}
}

func (converter *hecConverter) toReportRequest(
	reporterID uint64,
	attributes map[string]string,
	buffer *reportBuffer,
) []byte {
	return bytes.Join(converter.toSpans(buffer, attributes), []byte("\n"))
}

// func (converter *hecConverter) toReporter(reporterID uint64, attributes map[string]string) *collectorpb.Reporter {
// 	return &collectorpb.Reporter{
// 		ReporterId: reporterID,
// 		Tags:       converter.toFields(attributes),
// 	}
// }

func (converter *hecConverter) toSpans(buffer *reportBuffer, attributes map[string]string) [][]byte {
	spans := make([][]byte, len(buffer.rawSpans))
	for i, span := range buffer.rawSpans {
		spans[i] = converter.toSpan(span, buffer, attributes)
	}
	return spans
}

func (converter *hecConverter) toSpan(span RawSpan, buffer *reportBuffer, attributes map[string]string) []byte {
	span_map := make(map[string]interface{})
	if span.ParentSpanID == 0 {
		span_map["parent_span_id"] = nil
	} else {
		span_map["parent_span_id"] = strconv.FormatUint(span.ParentSpanID, 16)
	}
	span_map["trace_id"] 				= strconv.FormatUint(span.Context.TraceID, 16)
	span_map["span_id"] 				= strconv.FormatUint(span.Context.SpanID, 16)
	span_map["operation_name"] 			= span.Operation
	// span_map["parent_span_id"] 			= &psi// span.ParentSpanID
	span_map["timestamp"]				= converter.toTimestamp(span.Start)
	span_map["duration"] 				= converter.fromDuration(span.Duration)
	span_map["tags"] 					= make(map[string]interface{})
	span_map["baggage"] 				= &span.Context.Baggage
	// span_map["Logs"] 					= converter.toLogs(span.Logs, buffer)

	for key, value := range attributes {
		if strings.HasPrefix(key, "tracer_") || key == "device" || key == "component_name" {
			span_map[key] = value
		} else {
			span_map["tags"].(map[string]interface{})[key] = value
		}
	}
	for key, value := range span.Tags {
			span_map["tags"].(map[string]interface{})[key] = value
	}
	span_thing := make(map[string]interface{})
	span_thing["time"] = converter.toTimestamp(span.Start)
	span_thing["sourcetype"] = "splunktracing:span"
	span_thing["event"] = span_map		
	span_buffer, _ := json.Marshal(span_thing)

	// report_objs := make([][]byte, 1) 
	report_objs := make([][]byte, len(span.Logs) + 1)
	report_objs[0] = span_buffer
	//
	for idx, record := range span.Logs {
		log_map := make(map[string]interface{})
		log_map["timestamp"] = converter.toTimestamp(record.Timestamp)
		for k, v := range span_map {
			if k!="duration" && k!="timestamp" {
		    	log_map[k] = v
		    }
		}
		log_thing := make(map[string]interface{})
		log_thing["time"] = converter.toTimestamp(record.Timestamp)
		log_thing["sourcetype"] = "splunktracing:log"
		marshalFields(converter, log_map, record.Fields)
		log_thing["event"] = log_map
		log_buffer, _ := json.Marshal(log_thing)
		report_objs[idx+1] = log_buffer
	}
	//
	return bytes.Join(report_objs, []byte("\n"))
}

// func (converter *hecConverter) toLogs(records []opentracing.LogRecord, buffer *reportBuffer) []*collectorpb.Log {
// 	logs := make([]*collectorpb.Log, len(records))
// 	for i, record := range records {
// 		logs[i] = converter.toLog(record, buffer)
// 	}
// 	return logs
// }

// func (converter *hecConverter) toLog(record opentracing.LogRecord, buffer *reportBuffer) *collectorpb.Log {
// 	log := &collectorpb.Log{
// 		Timestamp: converter.toTimestamp(record.Timestamp),
// 	}
// 	marshalFields(converter, log, record.Fields, buffer)
// 	return log
// }

func (converter *hecConverter) toTimestamp(t time.Time) float64 {
	return float64(t.Unix()) + float64(t.Nanosecond())/1000000000
}

func (converter *hecConverter) fromDuration(d time.Duration) uint64 {
	return uint64(d / time.Microsecond)
}

func (converter *hecConverter) fromTimeRange(oldestTime time.Time, youngestTime time.Time) uint64 {
	return converter.fromDuration(youngestTime.Sub(oldestTime))
}
