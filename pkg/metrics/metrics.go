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
	prometheus.MustRegister(failedTests)
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
