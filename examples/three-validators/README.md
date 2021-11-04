# Three Validators Bridge Network

## Overview

The goal of `Three Validators Bridge Network` is to showcase how validators configured for similar HCS topics and threshold accounts operate simultaneously.
The network consists of **three** validator nodes, which process incoming transfers and events, and a **read-only node**, used only to monitor the network and the bridge itself.

Example uses `Docker Compose`.

## How to run?

1. Create a bridge configuration for Hedera using the [scripts](../../scripts/testnet/README.md).
2. Deploy EVM contracts from the Contracts [repository](https://github.com/LimeChain/hedera-eth-bridge-contracts/blob/main/README.md#scripts).
3. Set necessary [bridge](./bridge.yml) configuration for the nodes.
4. Set necessary configurations for [Alice](./alice/config), [Bob](./bob/config), [Carol](./carol/config) and [Dave](./dave/config).
5. Run `docker-compose up`
