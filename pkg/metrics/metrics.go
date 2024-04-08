package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/prometheus/common/expfmt"
	"log"
	"os"
)

const EnvPromAggregationGatewayUrl = "PROM_AGGREGATION_GATEWAY_URL"

// Define a Prometheus counter to count the number of failed tests
var (
	failedTests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "aswa_checks",
			Help: "Failed synthetic test.",
		},
		[]string{"app"},
	)
)

func init() {
	prometheus.MustRegister(failedTests)
}

// IncrementFailedTestsCounter increments the counter for a given app
func IncrementFailedTestsCounter(app string) {
	failedTests.WithLabelValues(app).Inc()
}

// PushMetrics pushes all collected metrics to the PAG.
func PushMetrics() error {
	textFormat := expfmt.NewFormat(expfmt.TypeTextPlain)
	pusher := push.New(getPromAggregationgatewayUrl(), "monitoring").
		Collector(failedTests).
		Format(textFormat)
	if err := pusher.Push(); err != nil {
		return err
	}
	return nil
}

// getPromAggregationgatewayUrl retrieves the pag url from environment variables.
func getPromAggregationgatewayUrl() string {
	promAggregationGatewayUrl := os.Getenv(EnvPromAggregationGatewayUrl)
	if promAggregationGatewayUrl == "" {
		log.Println("PROM_AGGREGATION_GATEWAY_URL is not set")
		return ""
	}
	return promAggregationGatewayUrl
}
