# Three Validators Bridge Network

## Overview

The goal of `Three Validators Bridge Network` is to showcase how validators configured
for similar HCS topics and custodial accounts operate simultaneously.
Example uses `Docker Compose`.

## How to run?

1. Run hedera-setup.go with privateKey, accountId and network as flags to generate the configurations

    `go run hedera-setup.go --privateKey=/your private key/ --accountId=/your account id/ --network=/previewnet|testnet|mainnet/ --members=/int, the count of the wanted bridge custodians/`

2. Set necessary configurations for [Alice](./alice/config/application.yml), [Bob](./bob/config/application.yml)
   and [Carol](./carol/config/application.yml)
3. Run `docker-compose up`
