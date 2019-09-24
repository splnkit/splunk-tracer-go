package splunktracing

// keys for splunk meta span events
const SPLMetaEvent_MetaEventKey = "splunk.meta_event"
const SPLMetaEvent_PropagationFormatKey = "splunk.propagation_format"
const SPLMetaEvent_TraceIdKey = "splunk.trace_id"
const SPLMetaEvent_SpanIdKey = "splunk.span_id"
const SPLMetaEvent_TracerGuidKey = "splunk.tracer_guid"

// operation names for splunk meta span events
const SPLMetaEvent_ExtractOperation = "splunk.extract_span"
const SPLMetaEvent_InjectOperation = "splunk.inject_span"
const SPLMetaEvent_SpanStartOperation = "splunk.span_start"
const SPLMetaEvent_SpanFinishOperation = "splunk.span_finish"
const SPLMetaEvent_TracerCreateOperation = "splunk.tracer_create"
