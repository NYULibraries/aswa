# ASWA (Application Status Watch Agent)

ASWA is a specialized monitoring tool designed to perform HTTP-based health checks on a variety of web endpoints.
It is engineered to ping specified URLs and validate their HTTP status codes and content against pre-configured expectations.
This enables real-time, automated verification of service availability and data integrity.

## Usage

ASWA supports three types of application environments:

* Development (dev): For testing and development purposes.
* Production (prod): For monitoring live, production-level services.
* Software-as-a-Service (saas): For third-party or cloud-based services.

You can specify which config file to load (dev, prod, saas) by setting the `YAML_PATH` environment variable. 
If no config file is specified, it will default to `dev.applications.yml`

Run a synthetic test in a docker container:

```
docker-compose run aswa $APP_NAME

docker-compose run aswa

```

Run a synthetic test locally:

```
./aswa $APP_NAME

./aswa 
```

### YAML config

Name, url, expected status, timeout and expected location of an application are specified in config/applications.yml

~~~ {.yml}
applications:
  - name: specialcollections
    url: 'https://specialcollections.library.nyu.edu/search/'
    expected_status: 200
    timeout: 600ms
~~~
