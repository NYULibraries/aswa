# ASWA (Application Status Watch Agent)

## Usage

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

