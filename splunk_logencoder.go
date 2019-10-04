package splunktracing

import (
	"encoding/json"
	// "fmt"

	"github.com/opentracing/opentracing-go/log"
)

const (
	ellipsis = "…"
)

// An implementation of the log.Encoder interface
type splunkLogFieldEncoder struct {
	converter *hecConverter
	keyValues map[string]interface{}
}

func marshalFields(
	converter *hecConverter,
	protoLog map[string]interface{},
	fields []log.Field,
) {
	logFieldEncoder := splunkLogFieldEncoder{
		converter: converter,
		keyValues: make(map[string]interface{}),
	}
	for _, field := range fields {
		field.Marshal(&logFieldEncoder)
	}
	protoLog["fields"] = logFieldEncoder.keyValues
}

func (lfe *splunkLogFieldEncoder) EmitString(key, value string) {
	safeKey := lfe.setSafeKey(key)
	safeValue := lfe.setSafeStringValue(value)
	lfe.keyValues[safeKey] = safeValue
}

func (lfe *splunkLogFieldEncoder) EmitBool(key string, value bool) {
	safeKey := lfe.setSafeKey(key)
	lfe.keyValues[safeKey] = value
}

func (lfe *splunkLogFieldEncoder) EmitInt(key string, value int) {
	safeKey := lfe.setSafeKey(key)
	lfe.keyValues[safeKey] = value
}

func (lfe *splunkLogFieldEncoder) EmitInt32(key string, value int32) {
	safeKey := lfe.setSafeKey(key)
	lfe.keyValues[safeKey] = value
}

func (lfe *splunkLogFieldEncoder) EmitInt64(key string, value int64) {
	safeKey := lfe.setSafeKey(key)
	lfe.keyValues[safeKey] = value
}

// N.B. We are using a string encoding for 32- and 64-bit unsigned
// integers because it will require a protocol change to treat this
// properly. Revisit this after the OC/OT merger.  LS-1175
//
// We could safely continue using the int64 value to represent uint32
// without breaking the stringified representation, but for
// consistency with uint64, we're encoding all unsigned integers as
// strings.
func (lfe *splunkLogFieldEncoder) EmitUint32(key string, value uint32) {
	safeKey := lfe.setSafeKey(key)
	lfe.keyValues[safeKey] = value
}

func (lfe *splunkLogFieldEncoder) EmitUint64(key string, value uint64) {
	safeKey := lfe.setSafeKey(key)
	lfe.keyValues[safeKey] = value
}

func (lfe *splunkLogFieldEncoder) EmitFloat32(key string, value float32) {
	safeKey := lfe.setSafeKey(key)
	lfe.keyValues[safeKey] = value
}

func (lfe *splunkLogFieldEncoder) EmitFloat64(key string, value float64) {
	safeKey := lfe.setSafeKey(key)
	lfe.keyValues[safeKey] = value
}

func (lfe *splunkLogFieldEncoder) EmitObject(key string, value interface{}) {
	safeKey := lfe.setSafeKey(key)
	jsonBytes, err := json.Marshal(value)
	if err != nil {
		emitEvent(newEventUnsupportedValue(key, value, err))
		safeValue := lfe.setSafeStringValue("<json.Marshal error>")
		lfe.keyValues[safeKey] = safeValue
		return
	}
	lfe.keyValues[safeKey] = lfe.setSafeJSONValue(string(jsonBytes))
}
func (lfe *splunkLogFieldEncoder) EmitLazyLogger(value log.LazyLogger) {
	// Delegate to `value` to do the late-bound encoding.
	value(lfe)
}

func (lfe *splunkLogFieldEncoder) setSafeStringValue(str string) string {
	if lfe.converter.maxLogValueLen > 0 && len(str) > lfe.converter.maxLogValueLen {
		str = str[:(lfe.converter.maxLogValueLen-1)] + ellipsis
	}
	return str
}

func (lfe *splunkLogFieldEncoder) setSafeJSONValue(json string) string{
	if lfe.converter.maxLogValueLen > 0 && len(json) > lfe.converter.maxLogValueLen {
		str := json[:(lfe.converter.maxLogValueLen-1)] + ellipsis
		return str
	}
	return json
}

func (lfe *splunkLogFieldEncoder) setSafeKey(key string) string {
	if lfe.converter.maxLogKeyLen > 0 && len(key) > lfe.converter.maxLogKeyLen {
		key = key[:(lfe.converter.maxLogKeyLen-1)] + ellipsis
	}
	return key
}