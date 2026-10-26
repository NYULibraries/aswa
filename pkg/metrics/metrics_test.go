package metrics

import (
	"context"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"slices"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	c "github.com/NYULibraries/aswa/pkg/config"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func TestPushMetrics(t *testing.T) {
	tests := []struct {
		name          string
		mockResponse  func(w http.ResponseWriter, r *http.Request)
		expectedError bool
	}{
		{
			name: "Successful push",
			mockResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			expectedError: false,
		},
		{
			name: "Failed push with non-200 response",
			mockResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock HTTP server for this test case
			server := httptest.NewServer(http.HandlerFunc(tt.mockResponse))
			defer server.Close()

			// Setup environment variable to mock the pushgateway URL
			t.Setenv(c.EnvPromAggregationGatewayUrl, server.URL)

			// Increment test counters to simulate metrics that would be pushed
			IncrementFailedTestsCounter("testApp")
			IncrementRunCounter("testApp")

			// Call PushMetrics to attempt to push the test counter to the mock server
			err := PushMetrics()

			// Assert that an error occurred only if one was expected
			if tt.expectedError {
				if err == nil {
					t.Errorf("TestPushMetrics %s: expected an error but got none", tt.name)
				}
			} else {
				if err != nil {
					t.Errorf("TestPushMetrics %s: did not expect an error but got: %v", tt.name, err)
				}
			}
		})
	}
}

// TestPushMetricsRequest pins down what reaches the Prom Aggregation Gateway: the method, path,
// text format, and both metric families with their env and app labels. Dashboards and alerts
// depend on these, and a Prometheus client library upgrade could change them.
func TestPushMetricsRequest(t *testing.T) {
	var method, path, contentType, body string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method, path, contentType = r.Method, r.URL.Path, r.Header.Get("Content-Type")
		b, _ := io.ReadAll(r.Body)
		body = string(b)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	t.Setenv(c.EnvPromAggregationGatewayUrl, server.URL)
	t.Setenv(c.EnvName, "format-test")

	IncrementRunCounter("push-format-app")
	IncrementFailedTestsCounter("push-format-app")
	// The counters are package globals, so read their values instead of assuming 1 (go test -count=2).
	runs := counterValue(t, checksRun.WithLabelValues("format-test", "push-format-app"))
	failures := counterValue(t, failedTests.WithLabelValues("format-test", "push-format-app"))

	if err := PushMetrics(); err != nil {
		t.Fatalf("PushMetrics: %v", err)
	}

	if method != http.MethodPut {
		t.Errorf("method = %q, want PUT", method)
	}
	if path != "/metrics/job/monitoring" {
		t.Errorf("path = %q, want /metrics/job/monitoring", path)
	}
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil || mediaType != "text/plain" || params["version"] != "0.0.4" {
		t.Errorf("Content-Type = %q, want text/plain; version=0.0.4", contentType)
	}
	// Match whole lines, so a wrong value such as 10 can't pass for 1.
	lines := make(map[string]bool)
	for line := range strings.SplitSeq(body, "\n") {
		lines[line] = true
	}
	for _, want := range []string{
		"# TYPE aswa_checks_run_total counter",
		"# TYPE aswa_checks_failed_total counter",
		fmt.Sprintf(`aswa_checks_run_total{app="push-format-app",env="format-test"} %g`, runs),
		fmt.Sprintf(`aswa_checks_failed_total{app="push-format-app",env="format-test"} %g`, failures),
	} {
		if !lines[want] {
			t.Errorf("pushed body is missing the line %q", want)
		}
	}
}

// stubSleep replaces the retry delays and the sleep between push attempts for one test, so
// retries don't wait. It records each requested wait and calls onSleep (if set) with the
// 1-based number of the wait.
func stubSleep(t *testing.T, delays []time.Duration, onSleep func(call int)) *[]time.Duration {
	t.Helper()
	origSleep, origDelays := sleep, pushRetryDelays
	t.Cleanup(func() { sleep, pushRetryDelays = origSleep, origDelays })
	pushRetryDelays = delays
	waits := &[]time.Duration{}
	sleep = func(d time.Duration) {
		*waits = append(*waits, d)
		if onSleep != nil {
			onSleep(len(*waits))
		}
	}
	return waits
}

// stubDial makes the HTTP client refuse connections while refuse returns true, the way a Service
// with no ready PAG pod does, and counts the refused attempts. It swaps http.DefaultTransport,
// which the pusher's client uses, for this test only, so no test port is released and rebound.
func stubDial(t *testing.T, refuse func() bool) *atomic.Int32 {
	t.Helper()
	orig := http.DefaultTransport
	transport := orig.(*http.Transport).Clone()
	dial := transport.DialContext
	refused := &atomic.Int32{}
	transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		if refuse() {
			refused.Add(1)
			return nil, &net.OpError{Op: "dial", Net: network, Err: os.NewSyscallError("connect", syscall.ECONNREFUSED)}
		}
		return dial(ctx, network, addr)
	}
	http.DefaultTransport = transport
	t.Cleanup(func() {
		http.DefaultTransport = orig
		transport.CloseIdleConnections()
	})
	return refused
}

// countingServer starts a PAG stand-in that answers every push with 200 and counts them.
func countingServer(t *testing.T) (*httptest.Server, *atomic.Int32) {
	t.Helper()
	pushes := &atomic.Int32{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pushes.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)
	return server, pushes
}

// TestPushMetricsRetriesUntilGatewayIsUp reproduces a PAG pod being rescheduled: connections are
// refused until the gateway comes back during the second wait, and the next attempt succeeds.
func TestPushMetricsRetriesUntilGatewayIsUp(t *testing.T) {
	server, pushes := countingServer(t)
	t.Setenv(c.EnvPromAggregationGatewayUrl, server.URL)

	var down atomic.Bool
	down.Store(true)
	refused := stubDial(t, down.Load)
	waits := stubSleep(t, []time.Duration{time.Second, 2 * time.Second, 3 * time.Second}, func(call int) {
		if call == 2 {
			down.Store(false)
		}
	})

	IncrementRunCounter("retry-app")
	if err := PushMetrics(); err != nil {
		t.Fatalf("PushMetrics: %v", err)
	}
	if got := refused.Load(); got != 2 {
		t.Errorf("refused %d connections, want 2", got)
	}
	if got := pushes.Load(); got != 1 {
		t.Errorf("gateway received %d pushes, want 1", got)
	}
	if want := []time.Duration{time.Second, 2 * time.Second}; !slices.Equal(*waits, want) {
		t.Errorf("waits = %v, want %v", *waits, want)
	}
}

// TestPushMetricsGivesUpWhenGatewayStaysDown checks that every delay is used once, each wait is
// followed by another attempt, and the connection error is returned, so the run still exits non-zero.
func TestPushMetricsGivesUpWhenGatewayStaysDown(t *testing.T) {
	server, pushes := countingServer(t)
	t.Setenv(c.EnvPromAggregationGatewayUrl, server.URL)
	refused := stubDial(t, func() bool { return true })
	delays := []time.Duration{time.Second, 2 * time.Second}
	waits := stubSleep(t, delays, nil)

	err := PushMetrics()
	if err == nil || !isConnectionError(err) {
		t.Fatalf("PushMetrics error = %v, want a connection error", err)
	}
	if got, want := refused.Load(), int32(len(delays)+1); got != want {
		t.Errorf("refused %d connections, want %d", got, want)
	}
	if got := pushes.Load(); got != 0 {
		t.Errorf("gateway received %d pushes, want 0", got)
	}
	if !slices.Equal(*waits, delays) {
		t.Errorf("waits = %v, want %v", *waits, delays)
	}
}

// TestPushMetricsDoesNotRetryRejectedPush checks that a push the PAG received and answered with
// an error is not sent again, since the PAG may already have added its counters.
func TestPushMetricsDoesNotRetryRejectedPush(t *testing.T) {
	var pushes atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pushes.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()
	t.Setenv(c.EnvPromAggregationGatewayUrl, server.URL)
	waits := stubSleep(t, []time.Duration{time.Second}, nil)

	if err := PushMetrics(); err == nil {
		t.Fatal("PushMetrics: expected an error for a 500 response")
	}
	if got := pushes.Load(); got != 1 {
		t.Errorf("gateway received %d pushes, want 1", got)
	}
	if len(*waits) != 0 {
		t.Errorf("waits = %v, want none", *waits)
	}
}

// collectedSeries counts the series a collector currently holds, without creating any
// (unlike WithLabelValues, which would add the child it looks up).
func collectedSeries(coll prometheus.Collector) int {
	ch := make(chan prometheus.Metric)
	go func() {
		coll.Collect(ch)
		close(ch)
	}()
	n := 0
	for range ch {
		n++
	}
	return n
}

func counterValue(t *testing.T, counter prometheus.Counter) float64 {
	t.Helper()
	var m dto.Metric
	if err := counter.Write(&m); err != nil {
		t.Fatalf("write counter: %v", err)
	}
	return m.GetCounter().GetValue()
}

// TestIncrementRunCounterInitializesFailureSeries verifies that recording a run for an app
// that has never failed still creates its aswa_checks_failed_total series at 0, so the uptime
// formula returns a value (100%) for always-healthy apps instead of no series at all.
func TestIncrementRunCounterInitializesFailureSeries(t *testing.T) {
	const app = "never-fails-unique"

	failuresBefore := collectedSeries(failedTests)

	IncrementRunCounter(app)

	if got := collectedSeries(failedTests); got != failuresBefore+1 {
		t.Errorf("expected IncrementRunCounter to create the failure series for %q (count %d -> %d)", app, failuresBefore, got)
	}

	env := c.GetEnvironmentName()
	if v := counterValue(t, failedTests.WithLabelValues(env, app)); v != 0 {
		t.Errorf("expected failure counter for %q to be initialized to 0 (not incremented), got %v", app, v)
	}
	if v := counterValue(t, checksRun.WithLabelValues(env, app)); v != 1 {
		t.Errorf("expected run counter for %q to be 1, got %v", app, v)
	}
}
