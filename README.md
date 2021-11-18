
<div align="center">

# Hashport Validator


[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
![Go build](https://github.com/LimeChain/hedera-eth-bridge-validator/workflows/Go%20build/badge.svg)
![Go Test](https://github.com/LimeChain/hedera-eth-bridge-validator/workflows/Go%20Test/badge.svg)
![E2E Tests](https://github.com/LimeChain/hedera-eth-bridge-validator/workflows/E2E%20Tests/badge.svg?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/LimeChain/hedera-eth-bridge-validator)](https://goreportcard.com/report/github.com/LimeChain/hedera-eth-bridge-validator)

</div>

## Overview 
This repository contains the Hedera <-> EVM Bridge Node. The bridge is operated by a set of validators who are running the Bridge Node software.

## Technologies
The Validator node is using Hedera Consensus Service for aggregating authorisation signatures resolving the need for nodes to have p2p communication and providing traceability for the bridging operations.
The node is a Go service with several watchers and handlers for Transfers, Message submission and EVM-based events.
Postgres is used for persisting state.

## Prerequisite Tools

Necessary tools prior to running the validator:

- [Docker](https://www.docker.com/products/docker-desktop)

## How to run?

To run the validator, execute the following commands in your terminal:

```
git clone https://github.com/LimeChain/hedera-eth-bridge-validator.git
cd hedera-eth-bridge-validator
docker-compose up
```

## Documentation
 - [Overview](docs/overview.md)
 - [Integration](docs/integration.md)
 - [Installation](docs/installation.md)
 - [Configuration](docs/configuration.md)
 - [Testing](docs/testing.md)
 - [Workflows](docs/workflows.md)
 - [Release](docs/release.md)

## Examples
* [Three Validators Bridge Network](./examples/three-validators/README.md)

## Mainnet Deployment

### Addresses/IDs

#### Ethereum
- Deployer Address: `0xD447148DB4AA5079113Cd0b16505A5CE3d4b62d1`
-  MultiSig: `0x10AB2d9C085c816f43ccB0FB1b27d024224E36f1`
- Router address:  `0x367e59b559283C8506207d75B0c5D8C66c4Cd4B7`
- OwnershipFacet address: `0x3fad3c29973bea5e964d9d90ebb8c84cb921e6c8`
- GovernanceFacet address: `0x48b3d6e97a8237f51861afb7f6512fb85a52d7ee`
- RouterFacet address:  `0xf9fe427563b12ec644e79a42b68e148273942b34`
- FeeCalculatorFacet address: `0x2bf0c79ce13a405a1f752f77227a28eec621f94c`
- DiamondCutFacet address: `0xe79c9bb508f00ab7b03cf3043929639e86626ef9`
- DiamondLoupeFacet address: `0x6f1fb462c6e328e8acccf59e58445a2fe18ff01e`
- PausableFacet address: `0x11481c1136d42c60c5bf29dfb9cb7eed90845814`

#### Polygon
- Deployer Address: `0x3E7056E7ff80969Fc86E1CD4871986F1E1126f8f`
- MultiSig: `0x40BF8172da97e9BeEA04Fb8C14836ABDdf46f3fb`
- Router address:  `0xf4C0153A8bdB3dfe8D135cE3bE0D45e14e5Ce53D`
- OwnershipFacet address:  `0x590243Fa41Af4383237E83a4CE5490a5AD9DacE3`
- GovernanceFacet address:  `0x8088Cb9ba08224c7Ecff05d4b9EE32DCAac1Fabc`
- RouterFacet address:  `0xA2F8f68d5d83f90b8401990196D0c233Dc0D4D7F`
- FeeCalculatorFacet address:  `0x9010EE70EC5d75Be46Ba5f7366776A3C7ad9Ab1f`
- DiamondCutFacet address:  `0x2232a10986375fdc9315F682551E141FC2A0a785`
- DiamondLoupeFacet address:  `0x8e1E4560C1571E4aBe2Be524CE62FE398bF8CAAD`
- PausableFacet address:  `0x9E4EAbD511acf7DC7594caBff98120139f9A43e1`

#### Hedera
- Deployer Account: `0.0.540219`
- Omnibus Account: `0.0.540219`
- Fee Account: `0.0.540286`
- HCS Topic: `0.0.540282`