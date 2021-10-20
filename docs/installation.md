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

## Running via Docker Compose

Docker Compose consists of scripts for the following components
 - PostgreSQL database
 - Hedera-Ethereum Bridge Validator (the application itself)
 
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
