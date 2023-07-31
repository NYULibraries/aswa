# ASWA (Application Status Watch Agent)

## Usage

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
