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
	// Try to register against the default registerer. Any errors are logged
	// (we intentionally don't panic during package init).
	_ = RegisterMetrics(prometheus.DefaultRegisterer)
}

// RegisterMetrics attempts to register the package-level collectors with the
// provided Registerer. If a collector with the same fully-qualified name is
// already registered, it will reuse the existing collector (to avoid
// duplicate registration errors). Returns any unexpected registration error.
func RegisterMetrics(reg prometheus.Registerer) error {
	if err := reg.Register(failedTests); err != nil {
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			if existing, ok := are.ExistingCollector.(*prometheus.CounterVec); ok {
				// Reuse the already-registered collector
				failedTests = existing
				return nil
			}
		}
		// Unexpected error: log it and return so callers/tests can observe it.
		log.Printf("prometheus: failed to register collector: %v", err)
		return err
	}
	return nil
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
