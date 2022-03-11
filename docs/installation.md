# Installation

The Hedera-Ethereum Bridge Validator can be ran [locally](#local-development) or via [Docker](#running-via-docker-compose).
To run locally, it will need Go. To run via Docker, either build locally, or pull the latest images from GCR:

```
docker-compose pull
```

**_NOTE:_** All commands run from the default directory of the repository.

## Prerequisites
- [Go 1.13+](https://golang.org/doc/install)
- [docker](https://docs.docker.com/install/)

## Local development

### Database setup

In addition to Go, you will need to run [PostgreSQL](https://postgresql.org) v9.6 database. The easiest way you can do that is by running it from the docker image:

```shell
docker-compose up db
```

This would start the postgres container and expose it to port 5432 with the default values configured.
The folder `./db` will be used mounting the volume and persisting the data.

You can run the database separately, but you will have to edit the validator default configuration for the database name, user and password.

### Build application

```shell
go build -o node cmd/*
```

### Run application

After you have run the database and compiled the node, you need to have the necessary [configuration](configuration.md) populated and run:
```shell
./node
```

### Unit Tests
In order to run the unit tests, one must execute the following command:
```shell
go test $(go list ./... | grep -v e2e)
```
The command filters out the e2e tests that require an additional setup.

To get the coverage report execute the following command:
```shell
go test -cover $(go list ./... | grep -v e2e)
```

### Monitoring
In order to start the Prometheus:
```shell
docker-compose up prometheus
```
To start the Grafana:
```shell
docker-compose up grafana
```
## Running via Docker Compose

Docker Compose consists of scripts for the following components
 - PostgreSQL database
 - Hedera-Ethereum Bridge Validator (the application itself)
 - Prometheus
 - Grafana
 
Containers use the following persisted volumes:
 - `./db` on your local machine maps to `/var/lib/postgresql/data` in the container, containing all necessary files
   for the PostgreSQL database. If the database container fails to initialize properly, or the database fails to run,
   it will be more likely to have to delete this folder. 

### How to run?

Before you run, [configure](configuration.md) the application updating the [node.yml](../node.yml) and [bridge.yml](../bridge.yml) file configuration. 
This file persists as a volume to the `Application` container.

Finally, run:
```shell
docker-compose up
```

Shutting down the containers:
```shell
docker-compose down
```

### Monitoring
Prometheus and Grafana are used for monitoring the Hashport. The monitoring is part of the Docker Compose setup.
By default, the Prometheus service is on port `9090`. The Grafana one is on port `3000`.
- The Prometheus is using `prometheus.yml` where the scraping configuration is set. The metrics path is `/api/v1/metrics`.
  The target setup is `<VALIDATOR_SERVICE_NAME>:<CONTAINER_PORT>`. For examples: `validator:5200` and `alice:5200`.
- The Prometheus service is set as the default data source for the Grafana service in `/grafana/datasource.yaml`.
- The default metrics path is set on `:9090/metrics`. The Prometheus UI is available on `:9090/graph`.
- The validator's metrics are on `:<VALIDATORS_HOST_PORT>/api/v1/metrics`.
For the `validator` example `:80/api/v1/metrics` and for the `alice` example `:6200/api/v1/metrics`
- The default credentials for Grafana are configurable through the `config-overrides.env` files in `./monitoring/grafana/` in the respective deployments folders.