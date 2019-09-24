// Code generated by counterfeiter. DO NOT EDIT.
package splunkfakes

import (
	"sync"

	splunktracing "github.com/splunk/splunk-tracer-go"
)

type FakeSpanRecorder struct {
	RecordSpanStub        func(splunktracing.RawSpan)
	recordSpanMutex       sync.RWMutex
	recordSpanArgsForCall []struct {
		arg1 splunktracing.RawSpan
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeSpanRecorder) RecordSpan(arg1 splunktracing.RawSpan) {
	fake.recordSpanMutex.Lock()
	fake.recordSpanArgsForCall = append(fake.recordSpanArgsForCall, struct {
		arg1 splunktracing.RawSpan
	}{arg1})
	fake.recordInvocation("RecordSpan", []interface{}{arg1})
	fake.recordSpanMutex.Unlock()
	if fake.RecordSpanStub != nil {
		fake.RecordSpanStub(arg1)
	}
}

func (fake *FakeSpanRecorder) RecordSpanCallCount() int {
	fake.recordSpanMutex.RLock()
	defer fake.recordSpanMutex.RUnlock()
	return len(fake.recordSpanArgsForCall)
}

func (fake *FakeSpanRecorder) RecordSpanArgsForCall(i int) splunktracing.RawSpan {
	fake.recordSpanMutex.RLock()
	defer fake.recordSpanMutex.RUnlock()
	return fake.recordSpanArgsForCall[i].arg1
}

func (fake *FakeSpanRecorder) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.recordSpanMutex.RLock()
	defer fake.recordSpanMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeSpanRecorder) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ splunktracing.SpanRecorder = new(FakeSpanRecorder)
