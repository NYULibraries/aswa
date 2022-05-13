# ASWA (Application Status Watch App)

## Usage

Run a synthetic test:

```
ASWA_EXPECTED_STATUS=200 ASWA_URL=https://library.nyu.edu docker-compose run cli
```

Specify a timeout:

```
ASWA_TIMEOUT=500s ASWA_EXPECTED_STATUS=200 ASWA_URL=https://library.nyu.edu docker-compose run cli
```
