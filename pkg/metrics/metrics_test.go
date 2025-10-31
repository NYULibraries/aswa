package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"

	c "github.com/NYULibraries/aswa/pkg/config"

	"github.com/prometheus/client_golang/prometheus/testutil"
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

			// Use unique app label per subtest to avoid metric collisions
			app := t.Name()

			// Increment a test counter to simulate metrics that would be pushed
			IncrementFailedTestsCounter(app)

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

func TestPushMetricsUnsupportedMediaType(t *testing.T) {
	// Server returns 415 Unsupported Media Type to simulate "wrong format"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnsupportedMediaType)
	}))
	defer server.Close()

	t.Setenv(c.EnvPromAggregationGatewayUrl, server.URL)
	app := t.Name()
	IncrementFailedTestsCounter(app)

	if err := PushMetrics(); err == nil {
		t.Fatalf("expected error pushing to server that returns 415, got nil")
	}
}

func TestPushMetricsServerClosed(t *testing.T) {
	// Start and immediately close the server to cause connection errors
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	url := server.URL
	server.Close()

	t.Setenv(c.EnvPromAggregationGatewayUrl, url)
	app := t.Name()
	IncrementFailedTestsCounter(app)

	if err := PushMetrics(); err == nil {
		t.Fatalf("expected connection error pushing to closed server, got nil")
	}
}

func TestIncrementFailedTestsCounterIncreases(t *testing.T) {
	app := t.Name()

	// Record current value, increment, and ensure it increased by 1.
	before := testutil.ToFloat64(failedTests.WithLabelValues(c.GetEnvironmentName(), app))
	IncrementFailedTestsCounter(app)
	after := testutil.ToFloat64(failedTests.WithLabelValues(c.GetEnvironmentName(), app))

	if after-before != 1 {
		t.Fatalf("expected counter to increase by 1, before=%v after=%v", before, after)
	}
}
