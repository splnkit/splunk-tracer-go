package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/splnkit/splunk-tracer-go"
	"github.com/opentracing/opentracing-go"
)

const (
	controlPath           = "/control"
	resultPath            = "/result"
	controllerPort        = 8088
	hecPort               = 8088
	controllerHost        = "localhost"
	controllerAccessToken = "ignored"
	logsSizeMax           = 1 << 20
)

var (
	logPayloadStr string
)

func fatal(x ...interface{}) {
	panic(fmt.Sprintln(x...))
}

func init() {
	lps := make([]byte, logsSizeMax)
	for i := 0; i < len(lps); i++ {
		lps[i] = 'A' + byte(i%26)
	}
	logPayloadStr = string(lps)
}

type control struct {
	Concurrent int // How many routines, threads, etc.

	// How much work to perform under one span
	Work int64

	// How many repetitions
	Repeat int64

	// How many amortized nanoseconds to sleep after each span
	Sleep time.Duration
	// How many nanoseconds to sleep at once
	SleepInterval time.Duration

	// How many bytes per log statement
	BytesPerLog int64
	NumLogs     int64

	// Misc control bits
	Trace   bool // Trace the operation.
	Exit    bool // Terminate the test.
	Profile bool // Profile this operation
}

type testClient struct {
	baseURL string
	tracer  opentracing.Tracer
}

func work(n int64) int64 {
	const primeWork = 982451653
	x := int64(primeWork)
	for n != 0 {
		x *= primeWork
		n--
	}
	return x
}

func (t *testClient) getURL(path string) []byte {
	resp, err := http.Get(t.baseURL + path)
	if err != nil {
		fatal("Bench control request failed: ", err)
	}
	if resp.StatusCode != 200 {
		fatal("Bench control status != 200: ", resp.Status, ": ", path)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fatal("Bench error reading body: ", err)
	}
	return body
}

func (t *testClient) loop() {
	for {
		body := t.getURL(controlPath)

		ctrl := control{}
		if err := json.Unmarshal(body, &ctrl); err != nil {
			fatal("Bench control parse error: ", err)
		}
		if ctrl.Exit {
			return
		}
		timing, flusht, sleeps, answer := t.run(&ctrl)
		t.getURL(fmt.Sprintf(
			"%s?timing=%.9f&flush=%.9f&s=%.9f&a=%d",
			resultPath,
			timing.Seconds(),
			flusht.Seconds(),
			sleeps.Seconds(),
			answer))
	}
}

func testBody(control *control) (time.Duration, int64) {
	var sleepDebt time.Duration
	var answer int64
	var totalSleep time.Duration
	for i := int64(0); i < control.Repeat; i++ {
		span := opentracing.StartSpan("span/test")
		answer = work(control.Work)
		for i := int64(0); i < control.NumLogs; i++ {
			span.LogEventWithPayload("testlog",
				logPayloadStr[0:control.BytesPerLog])
		}
		span.Finish()
		sleepDebt += control.Sleep
		if sleepDebt <= control.SleepInterval {
			continue
		}
		begin := time.Now()
		time.Sleep(sleepDebt)
		elapsed := time.Since(begin)
		sleepDebt -= elapsed
		totalSleep += elapsed
	}
	return totalSleep, answer
}

func (t *testClient) run(control *control) (time.Duration, time.Duration, time.Duration, int64) {
	if control.Trace {
		opentracing.InitGlobalTracer(t.tracer)
	} else {
		opentracing.InitGlobalTracer(opentracing.NoopTracer{})
	}
	conc := control.Concurrent
	runtime.GOMAXPROCS(conc)
	runtime.GC()
	runtime.Gosched()

	var sleeps time.Duration
	var answer int64

	beginTest := time.Now()
	if conc == 1 {
		s, a := testBody(control)
		sleeps += s
		answer += a
	} else {
		start := &sync.WaitGroup{}
		finish := &sync.WaitGroup{}
		start.Add(conc)
		finish.Add(conc)
		for c := 0; c < conc; c++ {
			go func() {
				start.Done()
				start.Wait()
				s, a := testBody(control)
				sleeps += s
				answer += a
				finish.Done()
			}()
		}
		finish.Wait()
	}
	endTime := time.Now()
	flushDur := time.Duration(0)
	if control.Trace {
		recorder, ok := t.tracer.(splunktracing.Tracer)
		if !ok {
			panic("Tracer does not have a splunk recorder")
		}
		recorder.Flush(context.Background())
		flushDur = time.Since(endTime)
	}
	return endTime.Sub(beginTest), flushDur, sleeps, answer
}

func main() {
	flag.Parse()
	tc := &testClient{
		baseURL: fmt.Sprint("http://",
			controllerHost, ":",
			controllerPort),
		tracer: splunktracing.NewTracer(splunktracing.Options{
			AccessToken: controllerAccessToken,
			Collector: splunktracing.Endpoint{
				Host:      controllerHost,
				Port:      hecPort,
				Plaintext: true,
			},
		}),
	}
	tc.loop()
}
