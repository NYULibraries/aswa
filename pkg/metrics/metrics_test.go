package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"

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
