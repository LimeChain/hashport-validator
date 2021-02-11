# Workflows

## Image release

Every time a new [release](https://github.com/LimeChain/hedera-eth-bridge-validator/releases) is published,
a Github Action workflow publishes a new image here [here](https://console.cloud.google.com/gcr/images/hedera-eth-bridge-test/GLOBAL/hedera-eth-bridge-validator).

## Testnet

When a new image publishes, a Github Action workflow deploys three validator applications in GCloud, simulating
a bridge application on Testnet.