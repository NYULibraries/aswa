package metrics

import (
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
	appInfo = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "aswa_app_info",
			Help: "Static application metadata for ASWA checks.",
		},
		[]string{"env", "app", "url"},
	)
)

func init() {
	prometheus.MustRegister(failedTests)
	prometheus.MustRegister(appInfo)
}

// IncrementFailedTestsCounter increments the counter for a given app.
func IncrementFailedTestsCounter(app string) {
	env := c.GetEnvironmentName()
	failedTests.WithLabelValues(env, app).Inc()
}

// SetAppInfo records static app metadata that can be joined into alerts.
func SetAppInfo(app string, url string) {
	env := c.GetEnvironmentName()
	appInfo.WithLabelValues(env, app, url).Set(1)
}

// PushMetrics pushes all collected metrics to the PAG.
func PushMetrics() error {
	textFormat := expfmt.NewFormat(expfmt.TypeTextPlain)
	pusher := push.New(c.GetPromAggregationgatewayUrl(), "monitoring").
		Collector(failedTests).
		Collector(appInfo).
		Format(textFormat)
	if err := pusher.Push(); err != nil {
		return err
	}
	return nil
}
