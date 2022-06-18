# ASWA (Application Status Watch Agent)

## Usage

Run a synthetic test in a docker container:

```
docker-compose run cli $APP_NAME
```

Run a synthetic test locally:

```
./aswa $APP_NAME
```

### YAML config

Name, url, expected status and expected location of an application are specified in config/applications.yml

Timeout is currently defaulting to `1*time.Minute`
~~~ {.yml}
applications:
  - name: specialcollections
    url: 'https://specialcollections.library.nyu.edu/search'
    expected_status: 200
    expectedLocation: 'https://specialcollections.library.nyu.edu/search'
~~~


Configuration of timeout duration in config YAML will be implemented in the next release