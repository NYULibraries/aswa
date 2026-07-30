package metrics

import (
	c "github.com/NYULibraries/aswa/pkg/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/prometheus/common/expfmt"
)

// Prometheus counters for synthetic checks:
//   - failedTests counts checks that FAILED.
//   - checksRun counts every check that RAN (pass or fail). It is the denominator that makes a
//     real uptime % computable: 1 - increase(aswa_checks_failed_total) / increase(aswa_checks_run_total).
var (
	failedTests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "aswa_checks_failed_total",
			Help: "Failed synthetic checks.",
		},
		[]string{"env", "app"},
	)

	checksRun = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "aswa_checks_run_total",
			Help: "Total synthetic checks run (pass or fail).",
		},
		[]string{"env", "app"},
	)
)

func init() {
	prometheus.MustRegister(failedTests, checksRun)
}

// IncrementFailedTestsCounter increments the failure counter for a given app.
func IncrementFailedTestsCounter(app string) {
	env := c.GetEnvironmentName()
	failedTests.WithLabelValues(env, app).Inc()
}

// IncrementRunCounter increments the run counter for a given app. Call it on every
// check (pass or fail) so uptime can be derived as 1 - failures/runs. It also ensures
// the app's failure series (aswa_checks_failed_total) exists, initialized to 0 when the app
// has never failed, so always-healthy apps still produce a series for the uptime formula
// 1 - increase(aswa_checks_failed_total) / increase(aswa_checks_run_total).
func IncrementRunCounter(app string) {
	env := c.GetEnvironmentName()
	checksRun.WithLabelValues(env, app).Inc()
	failedTests.WithLabelValues(env, app) // create the failure series at 0 if it does not exist yet
}

// PushMetrics pushes all collected metrics to the PAG.
func PushMetrics() error {
	textFormat := expfmt.NewFormat(expfmt.TypeTextPlain)
	pusher := push.New(c.GetPromAggregationgatewayUrl(), "monitoring").
		Collector(failedTests).
		Collector(checksRun).
		Format(textFormat)
	if err := pusher.Push(); err != nil {
		return err
	}
	return nil
}
