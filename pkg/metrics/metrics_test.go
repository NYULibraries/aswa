package metrics

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	c "github.com/NYULibraries/aswa/pkg/config"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

// NOTE: Do not use t.Parallel() in this file; tests mutate package-level globals.

// helper: assert that IncrementFailedTestsCounter increases by one for given labels
func assertIncByOne(t *testing.T, counter *prometheus.CounterVec, env, app string) {
	t.Helper()
	before := testutil.ToFloat64(counter.WithLabelValues(env, app))
	IncrementFailedTestsCounter(app)
	after := testutil.ToFloat64(counter.WithLabelValues(env, app))
	if after-before != 1 {
		t.Fatalf("expected +1; before=%v after=%v but got +%v", before, after, after-before)
	}
}

// helper: create a fresh CounterVec with the same name/labels as production one
func newFailedTestsCounter() *prometheus.CounterVec {
	return prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "aswa_checks_total",
			Help: "Failed synthetic test.",
		},
		[]string{"env", "app"},
	)
}

// helper: swap the package-level var for test isolation and restore on cleanup
func swapFailedTests(t *testing.T) *prometheus.CounterVec {
	t.Helper()
	orig := failedTests
	t.Cleanup(func() {
		// Clean up test mutations: unregister temp collectors and restore global
		prometheus.Unregister(failedTests)
		prometheus.Unregister(orig)
		failedTests = orig
	})
	failedTests = newFailedTestsCounter()
	// Make sure it's not registered anywhere yet
	prometheus.Unregister(failedTests)
	return orig
}

// badRegisterer is a helper registerer that simulates registration errors.
type badRegisterer struct{ err error }

func (b badRegisterer) Register(prometheus.Collector) error  { return b.err }
func (b badRegisterer) Unregister(prometheus.Collector) bool { return false }
func (b badRegisterer) MustRegister(cs ...prometheus.Collector) {
	if b.err != nil {
		panic(b.err)
	}
}
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
	app, env := t.Name(), c.GetEnvironmentName()

	assertIncByOne(t, failedTests, env, app)
}
func TestRegisterMetrics_FreshRegistry_Succeeds(t *testing.T) {
	swapFailedTests(t)

	reg := prometheus.NewRegistry()
	if err := RegisterMetrics(reg); err != nil {
		t.Fatalf("RegisterMetrics should succeed on fresh registry: %v", err)
	}

	// Prove counter works
	app, env := t.Name(), c.GetEnvironmentName()
	assertIncByOne(t, failedTests, env, app)
}

func TestRegisterMetrics_AlreadyRegistered_SameType_Reuses(t *testing.T) {
	swapFailedTests(t)

	reg := prometheus.NewRegistry()

	// Pre-register another *CounterVec with the same fq name/labels
	existing := newFailedTestsCounter()
	if err := reg.Register(existing); err != nil {
		t.Fatalf("pre-register existing counter: %v", err)
	}

	// Our call should detect AlreadyRegistered and reuse existing
	if err := RegisterMetrics(reg); err != nil {
		t.Fatalf("RegisterMetrics should not error when same-type is already registered: %v", err)
	}

	// Functional proof of reuse: increment via package API and observe delta on existing
	app, env := t.Name(), c.GetEnvironmentName()
	assertIncByOne(t, existing, env, app)
}

func TestRegisterMetrics_AlreadyRegistered_DifferentType_ReturnsError_ButPushStillWorks(t *testing.T) {
	swapFailedTests(t)

	reg := prometheus.NewRegistry()

	// Register a GaugeVec with the same name/labels to force a type mismatch
	gv := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "aswa_checks_total",
			Help: "Failed synthetic test.",
		},
		[]string{"env", "app"},
	)
	if err := reg.Register(gv); err != nil {
		t.Fatalf("pre-register gaugevec: %v", err)
	}

	// Should return an AlreadyRegisteredError we do NOT swallow (different type)
	if err := RegisterMetrics(reg); err == nil {
		t.Fatalf("expected AlreadyRegisteredError for different type, got nil")
	}

	// Even if not registered, we still can use and push our collector directly.
	app, env := t.Name(), c.GetEnvironmentName()
	assertIncByOne(t, failedTests, env, app)

	// Push should succeed because pusher.Collector(failedTests) writes this collector explicitly.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	t.Setenv(c.EnvPromAggregationGatewayUrl, srv.URL)
	if err := PushMetrics(); err != nil {
		t.Fatalf("push should succeed even with registry name collision of different type: %v", err)
	}

}
func TestRegisterMetrics_UnexpectedError_BubblesUp(t *testing.T) {
	swapFailedTests(t)

	// A fake registerer that returns a non-AlreadyRegistered error
	sentinel := errors.New("boom")
	if err := RegisterMetrics(badRegisterer{err: sentinel}); !errors.Is(err, sentinel) {
		t.Fatalf("expected unexpected error to bubble up; got %v", err)
	}
}

func TestRegisterMetrics_IdempotentOnSameRegistry(t *testing.T) {
	swapFailedTests(t)

	reg := prometheus.NewRegistry()

	// First register succeeds
	if err := RegisterMetrics(reg); err != nil {
		t.Fatalf("first RegisterMetrics failed: %v", err)
	}

	// Second register should gracefully handle AlreadyRegistered and return nil
	if err := RegisterMetrics(reg); err != nil {
		t.Fatalf("second RegisterMetrics should be idempotent (nil), got: %v", err)
	}

	// Functional check
	app, env := t.Name(), c.GetEnvironmentName()
	assertIncByOne(t, failedTests, env, app)
}
