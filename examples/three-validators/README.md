# Three Validators Bridge Network

## Overview

The goal of `Three Validators Bridge Network` is to showcase how validators configured
for similar HCS topics and custodial accounts operate simultaneously.
Example uses `Docker Compose`.

## How to run?

1. Run hedera-setup.go with prKey and accountId as flags to generate the configurations

    `go run hedera-setup.go -prKey=/your private key/ -accountId=/your account id/`

2. Set necessary configurations for [Alice](./alice/config/application.yml), [Bob](./bob/config/application.yml)
   and [Carol](./carol/config/application.yml)
3. Run `docker-compose up`
