# Hedera - Ethereum Bridge Node

<div align="center">

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
![Go build](https://github.com/LimeChain/hedera-eth-bridge-validator/workflows/Go%20build/badge.svg)
![Go Test](https://github.com/LimeChain/hedera-eth-bridge-validator/workflows/Go%20Test/badge.svg)
![E2E Tests](https://github.com/LimeChain/hedera-eth-bridge-validator/workflows/E2E%20Tests/badge.svg)
![GitHub release (latest by date)](https://img.shields.io/github/v/release/LimeChain/hedera-eth-bridge-validator)

</div>

## Overview 
This repository contains the Hedera <-> Ethereum Bridge Node. The bridge is operated by a set of validators who are running the Bridge Node software.  
The diagram below shows the operations involved for the Hedera -> Ethereum bridging solution.

<p align="center">

![Hedera-Ethereum-MVP](docs/images/hedera-eth-mvp.png "Hedera->Ethereum") 

</p>

## Technologies
The Validator node is using Hedera Consensus Service for aggregating authorisation signatures resolving the need for node to have p2p communication and providing traceability for the transfer.
The node is a Go Lang service with several watchers and handlers for Crypto Transfer, Message submission and Ethereum events.
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
 - [Installation](docs/installation.md)
 - [Configuration](docs/configuration.md)
 - [Testing](docs/testing.md)
 - [Workflows](docs/workflows.md)
 - [Release](docs/release.md)

## Examples
* [Three Validators Bridge Network](./examples/three-validators/README.md)