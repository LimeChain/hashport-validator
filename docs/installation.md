# Installation

The Hedera-Ethereum Bridge Validator can be ran [locally](#local-development) or via [Docker](#running-via-docker-compose).
To run locally, it will need Go. To run via Docker, either build locally, or pull the latest images from GCR:

```
docker-compose pull
```

## Prerequisites
``
Go 1.13+
``

## Local development

### Database setup

In addition to Go, you will need to install a database and initialize it.
The application is using [PostgreSQL](https://postgresql.org) v9.6.

Run initialization script located at `scripts/init.sql`:
```
psql postgres -f scripts/init.sql
```

### Run application

After you have run the database, you need to have the necessary [configuration](configuration.md) populated and run:
```
go run cmd/main.go
```

#### Run in debug mode

```
go run cmd/main.go -debug=true
```

#### Build application

```
go build cmd/main.go -o validator
```

#### Unit Tests

```
go test ./...
```

## Running via Docker Compose

Docker Compose consists of scrips of the following components:
 - PostgreSQL database
 - Hedera-ETH Bridge Validator (Application)
 
Containers use the following persisted volumes:
 - `./db` on your local machine maps to `/var/lib/postgresql/data` in the container, containing all necessary files
   for the PostgreSQL database. If the database container fails to initialize properly, or the database fails to run,
   it will be more likely to have to delete this folder. 
   
### How to run?

Before you run, [configure](configuration.md) the application updating the [application.myl](../application.yml)
file configuration. This file persists as a volume to the `Application` container.

Finally, run:
```
docker-compose up
```

Shutting down the containers:
```
docker-compose down
```