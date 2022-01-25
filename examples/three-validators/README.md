# Three Validators Bridge Network

## Overview

The goal of `Three Validators Bridge Network` is to showcase how validators configured for similar HCS topics and threshold accounts operate simultaneously.
The network consists of **three** validator nodes, which process incoming transfers and events, and a **read-only node**, used only to monitor the network and the bridge itself.

Example uses `Docker Compose`.

## How to run?

1. Create a bridge configuration for Hedera using the [scripts](../../scripts/README.md).
2. Using the reference and the scripts for the Smart Contracts from [repository](https://github.com/LimeChain/hedera-eth-bridge-contracts/blob/main/README.md#scripts). Do the following: 
   1. Deploy Diamond Router (with its facets)
   2. Deploy Native Token on Ethereum Ropsten testnet
   3. Update Native Token in Diamond Router on Ethereum Ropsten Testnet
   4. Deploy Wrapped Token on Polygon testnet for the newly created Native Token on Ethereum Ropsten Testnet
   5. Deploy Native Token on Polygon testnet
   6. Update Native Token in Diamond Router on Polygon Testnet
   7. Deploy Wrapped Token on Ethereum Ropsten testnet for the newly created Native Token on Polygon Testnet
   8. Deploy Wrapped Token for Hedera's HBAR on Polygon Testnet
   9. Deploy Wrapped Token for Hedera's HBAR on Ethereum Ropsten Testnet
   10. Deploy Wrapped Token for Hedera's newly created Native Token on Ethereum Ropsten Testnet
   11. Deploy Wrapped Token for Hedera's newly created Native Token on Polygon Testnet
3. Set necessary [bridge](./bridge.yml) configuration for the nodes.
4. Set necessary configurations for [Alice](./alice/config), [Bob](./bob/config), [Carol](./carol/config) and [Dave](./dave/config).
5. Run `docker-compose up`
