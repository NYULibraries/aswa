package metrics

import (
	u "github.com/NYULibraries/aswa/pkg/utils"
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

	clusterInfoGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "aswa_cluster_info",
			Help: "Cluster info.",
		},
		[]string{"cluster_info"},
	)
)

func init() {
	prometheus.MustRegister(failedTests)
	prometheus.MustRegister(clusterInfoGauge)
}

// IncrementFailedTestsCounter increments the counter for a given app
func IncrementFailedTestsCounter(app string) {
	env := u.GetEnvironmentName()
	failedTests.WithLabelValues(env, app).Inc()
}

// PushMetrics pushes all collected metrics to the PAG.
func PushMetrics() error {
	clusterInfo := u.GetClusterInfo()
	clusterInfoGauge.WithLabelValues(clusterInfo).Set(1)
	textFormat := expfmt.NewFormat(expfmt.TypeTextPlain)
	pusher := push.New(u.GetPromAggregationgatewayUrl(), "monitoring").
		Collector(failedTests).
		Collector(clusterInfoGauge).
		Format(textFormat)
	if err := pusher.Push(); err != nil {
		return err
	}
	return nil
}
