# ASWA (Application Status Watch Agent)

ASWA is a specialized monitoring tool designed to perform HTTP-based health checks on a variety of web endpoints.
It is engineered to ping specified URLs and validate their HTTP status codes and content against pre-configured expectations.
This enables real-time, automated verification of service availability and data integrity.

## Usage

ASWA supports three types of application environments:

* Development (dev): For testing and development purposes.
* Production (prod): For monitoring live, production-level services.
* Software-as-a-Service (saas): For third-party or cloud-based services.

Configuration files are stored in the `config` directory. You can specify which config file to load (dev, prod, saas) by setting the `YAML_PATH` environment variable. 
If no config file is specified, it will default to `dev.applications.yml`

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
* `expected_content`: A string to match against the content returned by the URL.
* `expected_location`: The expected final URL after all redirects, if any.
* `timeout`: The maximum time to wait for a response, in milliseconds.
* `expected_csp`: The expected Content Security Policy (CSP) header value.

~~~ {.yml}
applications:
  - name: specialcollections
    url: 'https://specialcollections.library.nyu.edu/search/'
    expected_status: 200
    timeout: 600ms
~~~

### Environment variables
In the `docker-compose.yml` file, you can configure the environment variables for the ASWA service. 
Here is an explanation of the key environment variables:

* ENV: Specifies the environment in which ASWA is running (default is `dev`).
* DEBUG_MODE: Enables or disables debug mode (default is false).
* CLUSTER_INFO: Includes cluster information in the output.
* OUTPUT_SLACK: If set to true, results are sent to Slack; otherwise, they are sent to PAG (default is `false`).
* PROM_AGGREGATION_GATEWAY_URL: URL for the Prom Aggregation Gateway.
* SLACK_WEBHOOK_URL: Slack webhook URL for notifications.
* YAML_PATH: Path to the YAML configuration file (default is `config/dev.applications.yml`).

### Notifications
ASWA can post the results of its checks to respective Slack channels (dev, prod, saas) based on the environment. To enable this feature, set the `SLACK_WEBHOOK_URL` environment variable with your Slack webhook URL.

By default, `OUTPUT_SLACK` is set to `false`. If `OUTPUT_SLACK` is set to `true`, the results are sent to Slack. If `OUTPUT_SLACK` is set to `false`, the results are sent to PAG (Prom Aggregation Gateway). PAG aggregates metrics for Prometheus and is similar in function to Pushgateway but includes metric aggregation capabilities.


### Deployment
ASWA is designed to run as a cron job in a Kubernetes (K8s) cluster. 
To include cluster information in the output, set the `CLUSTER_INFO` environment variable.

### Tests
To run the tests, execute the following command:

Run tests locally:
```shell
go test -cover ./...
```

Run tests in a docker container:
```shell
docker compose run test


