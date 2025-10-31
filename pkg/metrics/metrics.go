package metrics

import (
	"log"

	c "github.com/NYULibraries/aswa/pkg/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/prometheus/common/expfmt"
)

// Define a Prometheus counter to count the number of failed tests
var (
	failedTests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "aswa_checks_total",
			Help: "Failed synthetic test.",
		},
		[]string{"env", "app"},
	)
)

func init() {
	// Attempt to register the collector. If another collector with the same
	// fully-qualified name is already registered, reuse it instead of
	// panicking. This makes tests and multiple-import scenarios resilient to
	// duplicate registration.
	if err := prometheus.Register(failedTests); err != nil {
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			if existing, ok := are.ExistingCollector.(*prometheus.CounterVec); ok {
				// Reuse the already-registered collector
				failedTests = existing
				return
			}
		}
		// Unexpected error: log it and continue. We avoid panicking in init so
		// package import does not bring down the process; callers can still
		// push metrics using the Collector directly even if registration
		// failed. This follows the project's no-panic policy.
		log.Printf("prometheus: failed to register collector: %v", err)
		return
	}
}

// IncrementFailedTestsCounter increments the counter for a given app
func IncrementFailedTestsCounter(app string) {
	env := c.GetEnvironmentName()
	failedTests.WithLabelValues(env, app).Inc()
}

// PushMetrics pushes all collected metrics to the PAG.
func PushMetrics() error {
	textFormat := expfmt.NewFormat(expfmt.TypeTextPlain)
	pusher := push.New(c.GetPromAggregationgatewayUrl(), "monitoring").
		Collector(failedTests).
		Format(textFormat)
	if err := pusher.Push(); err != nil {
		return err
	}
	return nil
}
