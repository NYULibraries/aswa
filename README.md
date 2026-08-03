# ASWA (Application Status Watch Agent)

ASWA is a specialized monitoring tool designed to perform HTTP-based health checks on a variety of web endpoints.
It is engineered to ping specified URLs and validate their HTTP status codes and content against pre-configured expectations.
This enables real-time, automated verification of service availability and data integrity.

## Usage

ASWA supports three types of application environments:

* Development (dev): For testing and development purposes.
* Production (prod): For monitoring live, production-level services.
* Software-as-a-Service (saas): For third-party or cloud-based services.

Configuration files are stored in the `config` directory. Set the `YAML_PATH` environment variable to choose one.
If it is unset, ASWA defaults to `config/dev.applications.yml`.

`YAML_PATH` is checked against a whitelist (`allowedConfigPaths` in `pkg/config/config.go`), so it must be one of these
exact **relative** paths. Anything else — including the absolute form `/config/prod.applications.yml` — is rejected with
`config file path is not allowed`:

* `config/dev.applications.yml`
* `config/primo_ve.applications.yml`
* `config/prod.applications.yml`
* `config/saas.applications.yml`

Adding a new config file therefore means adding it to `allowedConfigPaths` too. Setting `SKIP_WHITELIST_CHECK=true`
bypasses the check.

Run a synthetic test in a docker container:

```
docker compose run aswa $APP_NAME

docker compose run aswa

```

Run a synthetic test locally:

```
./aswa $APP_NAME

./aswa 
```

### Building ASWA binary
To build the ASWA binary, execute the following command:

```shell
go build
```

### YAML config

The configuration is defined in a YAML file and must adhere to the following schema:

Required Fields
* `name`: The name of the application, must be non-empty.
* `url`: The URL to ping, must be a valid URL and non-empty.
* `expected_status`: The expected HTTP status code, must be non-zero.

Optional Fields
* `expected_content`: A string to match against the content returned by the URL. Matching is a plain substring search
  over the body of the final landing page — setting this makes ASWA issue a second, redirect-following GET.
* `expected_location`: The expected `Location` header of the probe response — the next hop, not the final URL, since
  the probe does not follow redirects (see How a check runs). A relative value is compared on path and query only
  (a leading slash is optional); an absolute value is compared in full.
* `timeout`: The maximum time to wait for a response (Go duration string, e.g. `600ms`, `2s`). If unset, no client
  timeout is applied.
* `include_actual_content_on_failure`: If true, an `expected_content` mismatch reports the whole response body (read
  up to a 10 MiB cap) rather than just the expected string. Only enable it for small, non-sensitive pages.
  `DEBUG_MODE=true` has the same effect for every check in the primo_ve config.
* `max_redirects`: Maximum number of redirects to follow when `expected_content` is set (default: 10).
* `expected_csp`: The expected Content Security Policy (CSP) header value. Compared exactly.

~~~ {.yml}
applications:
  - name: specialcollections
    url: 'https://specialcollections.library.nyu.edu/search/'
    expected_status: 200
    timeout: 600ms
~~~

### How a check runs

A check makes one or two HTTP requests, and the two treat redirects differently:

1. **Probe — always.** A `HEAD` on the configured URL, falling back to `GET` when the server answers `405` or the
   request errors. Redirects are **not** followed, so `expected_status`, `expected_location` and `expected_csp` are all
   evaluated against this first response.
2. **Content fetch — only when `expected_content` is set.** A second `GET` that **does** follow redirects, up to
   `max_redirects` (default 10). `expected_content` is matched against the body of the final landing page, read up to
   a 10 MiB cap.

### Environment variables
In the `docker-compose.yml` file, you can configure the environment variables for the ASWA service. 
Here is an explanation of the key environment variables:

* ENV: Specifies the environment in which ASWA is running (default is `dev`). It becomes the `env` label on the metrics below.
* DEBUG_MODE: Enables or disables debug mode (default is false).
* CLUSTER_INFO: Cluster name. Used only to prefix Slack messages — see Notifications.
* OUTPUT_SLACK: If set to true, results are sent to Slack; otherwise, they are sent to PAG (default is `false`).
* PROM_AGGREGATION_GATEWAY_URL: URL for the Prom Aggregation Gateway.
* SLACK_WEBHOOK_URL: Slack webhook URL. Read by `entrypoint.sh`, not by the binary — see Notifications.
* YAML_PATH: Path to the YAML configuration file (default is `config/dev.applications.yml`).
* SKIP_WHITELIST_CHECK: If true, skips the `YAML_PATH` whitelist check.

### Notifications

Slack posting is done by `entrypoint.sh`, not by the Go binary, so it only happens when ASWA runs in its container.
The binary writes results to stdout; the entrypoint captures that output and POSTs it to `SLACK_WEBHOOK_URL` only when
**both** of these hold:

* `OUTPUT_SLACK` is `true`, and
* the output contains the word `Failure` — passing runs are never posted.

If `CLUSTER_INFO` is set, the message is prefixed with its uppercased value; otherwise the prefix reads
`Unknown cluster: CLUSTER_INFO is not set`. Running the binary directly (`./aswa`) never posts to Slack, whatever
`SLACK_WEBHOOK_URL` is set to.

By default, `OUTPUT_SLACK` is `false` and results go to PAG (Prom Aggregation Gateway) instead. PAG aggregates metrics
for Prometheus and is similar in function to Pushgateway but includes metric aggregation capabilities. The two modes are
mutually exclusive: with `OUTPUT_SLACK=true` no metrics are recorded or pushed at all.

### Metrics
When `OUTPUT_SLACK` is `false`, ASWA pushes two counters to PAG on **every** run:

* `aswa_checks_failed_total{env, app}` — incremented when a synthetic check **fails**.
* `aswa_checks_run_total{env, app}` — incremented on **every** check (pass or fail).

Because the run counter is recorded even when everything passes, real per-application uptime can be computed in Prometheus/Grafana:

```promql
uptime % = 1 - increase(aswa_checks_failed_total[$range]) / increase(aswa_checks_run_total[$range])
```

PAG sums counters across pushes, so both metrics accumulate correctly over time.

### Deployment

ASWA is designed to run as a cron job in a Kubernetes (K8s) cluster.
Set `CLUSTER_INFO` to identify the cluster in Slack notifications.

The config files are baked into the image at build time (the `Dockerfile` copies `/app/config` to `/config`) — there is
no ConfigMap mount. A config change therefore only reaches a cluster once the image has been rebuilt and re-pulled.

### Tests

To run the tests, execute the following command:

Run tests locally:

```shell
go test -cover ./...
```

Run tests in a docker container:

```shell
docker compose run test
```

### Make targets

A `Makefile` wraps the common workflows:

```shell
make build        # go build -o aswa .
make run          # build, then run the binary
make run-app foo  # run the synthetic check for app "foo" in a container
make test         # run the tests in the aswa_test container
make container    # docker compose build
make format       # go fmt ./...
make lint         # golangci-lint (make lint-install to install it)
make staticcheck  # staticcheck (make staticcheck-install to install it)
make all          # format, lint, staticcheck, build images, test, run in container
make clean        # remove the binary and tear down compose images/volumes
```
