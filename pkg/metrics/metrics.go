package metrics

import (
	"errors"
	"log"
	"net"
	"time"

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

// pushRetryDelays are the waits between push attempts while the PAG can't be reached. They add
// up to just under 2 minutes, enough for the PAG pod to be rescheduled onto another node.
var pushRetryDelays = []time.Duration{5 * time.Second, 10 * time.Second, 20 * time.Second, 40 * time.Second, 40 * time.Second}

// sleep is replaced in tests so retries don't wait.
var sleep = time.Sleep

// PushMetrics pushes all collected metrics to the PAG. While the PAG can't be reached it retries
// after each of pushRetryDelays, and it returns the last error once they run out.
func PushMetrics() error {
	textFormat := expfmt.NewFormat(expfmt.TypeTextPlain)
	pusher := push.New(c.GetPromAggregationgatewayUrl(), "monitoring").
		Collector(failedTests).
		Collector(checksRun).
		Format(textFormat)
	err := pusher.Push()
	for _, delay := range pushRetryDelays {
		if err == nil || !isConnectionError(err) {
			return err
		}
		log.Printf("Metrics push failed, retrying in %v: %v", delay, err)
		sleep(delay)
		err = pusher.Push()
	}
	return err
}

// isConnectionError reports whether err means the push never reached the PAG (connection refused,
// DNS lookup failure, connect timeout). Only those are retried: the PAG adds up the counters of
// every push it receives, so retrying a push it may have received could count a run twice.
// This assumes the configured URL is the PAG itself. If it redirected, a dial error could come
// from the redirect target after the first server had already received the push.
func isConnectionError(err error) bool {
	var opErr *net.OpError
	return errors.As(err, &opErr) && opErr.Op == "dial"
}
