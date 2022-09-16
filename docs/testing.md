# Testing

## E2E Testing

Before you run E2E tests, you need to have a running application.

### Configuration

Configure properties, needed to run e2e tests in `e2e/setup/application.yml` and `e2e/setup/bridge.yml`.
**Keep in mind that most of the configuration needs to be the same as the application's**.

Configuration for `e2e/setup/application.yml`:

| Name                                  | Description                                                                                                                                                                                                       |
|---------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `hedera.bridge_account`               | The configured Bridge account. Validators listen for CRYPTOTRANSFER transactions, crediting that account.                                                                                                         |
| `hedera.dbs[].host`                   | The IP or hostname used to connect to the database.                                                                                                                                                               |
| `hedera.dbs[].name`                   | The name of the database.                                                                                                                                                                                         |
| `hedera.dbs[].password`               | The database password the processor uses to connect.                                                                                                                                                              |
| `hedera.dbs[].port`                   | The port used to connect to the database.                                                                                                                                                                         |
| `hedera.dbs[].username`               | The username the processor uses to connect to the database.                                                                                                                                                       |
| `hedera.members[]`                    | The Hedera account ids of the validators, to which their bridge fees will be sent (if Bridge accepts Hedera Tokens, associations with these tokens will be required). Used to assert balances after transactions. |
| `hedera.mirror_node.api_address`      | The Hedera Rest API root endpoint. Depending on the Hedera network type, this will need to be changed.                                                                                                            |
| `hedera.mirror_node.client_address`   | The HCS Mirror node endpoint. Depending on the Hedera network type, this will need to be changed.                                                                                                                 |
| `hedera.mirror_node.polling_interval` | How often (in seconds) the application will poll the mirror node for new transactions.                                                                                                                            |
| `hedera.network_type`                 | Which Hedera network to use. Can be either `mainnet`, `previewnet`, `testnet`.                                                                                                                                    |
| `hedera.sender.account`               | The account that will be sending assets through the bridge.                                                                                                                                                       |
| `hedera.sender.private_key`           | The private key for the account that will be sending assets through the bridge.                                                                                                                                   |
| `hedera.topic_id`                     | The configured Bridge Topic. Validators watch & submit signatures to that Topic.                                                                                                                                  |
| `evm[]`                               | The chain id of the EVM network. Used as a key for the following `node.clients.evm[i].*` configuration fields below.                                                                                              |
| `evm[].block_confirmations`           | The number of block confirmations to wait for before processing an event for the given EVM network.                                                                                                               |
| `evm[].node_url`                      | The endpoint of the node for the given EVM network.                                                                                                                                                               |
| `evm[].private_key`                   | The private key for the given EVM network.                                                                                                                                                                        |
| `tokens.whbar`                        | The native asset, which represents HBAR. Used as a bridged asset between the two networks.                                                                                                                        |
| `tokens.wtoken`                       | The native asset, which represents a Token on Hedera. Used as a bridged asset between the two networks.                                                                                                           |
| `tokens.evm_native_token`             | The address of the EVM native token, which will be used as a bridged asset to Hedera and other EVMs.                                                                                                              |
| `tokens.nft_token`                    | The Hedera TokenID of the token, used in NFT E2E tests.                                                                                                                                                           |
| `tokens.nft_serial_number`            | The Hedera Serial Number of the token, used in NFT E2E tests.                                                                                                                                                     |
| `validator_url`                       | The URL of the Validator node. Used for querying Metadata.                                                                                                                                                        |
| `scenario.expectedValidatorsCount`    | Test scenario option describing the expected number of collected signatures                                                                                                                                       |
| `scenario.firstEvmChainId`            | Test scenario option describing the first chain Id                                                                                                                                                                |
| `scenario.secondEvmChainId`           | Test scenario option describing the second (target) chain Id                                                                                                                                                      |

Configuration for `e2e/setup/bridge.yml`:

| Name                                                   | Default | Description                                                                                                                                                                                                                                                          |
|--------------------------------------------------------|---------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `bridge.topic_id`                                      | ""      | The topic id, which the validators will use to monitor and submit consensus messages to.                                                                                                                                                                             |
| `bridge.networks[i]`                                   | ""      | The id of the network. **`0` stands for Hedera**. Every other id must be the `chainId` of the EVM network. Used as a key for the following `bridge.networks[i].*` configuration fields below.                                                                        |
| `bridge.networks[i].bridge_account`                    | ""      | The account id validators use to monitor for incoming transfers. Applies only for network with id `0`. Also, serves as a distributor for Hedera transfers (validator fees and bridged amounts).                                                                      |
| `bridge.networks[i].payer_account`                     | ""      | The account id paying for Hedera transfers fees. Applies **only** for network with id `0`.                                                                                                                                                                           |
| `bridge.networks[i].members`                           | []      | The Hedera account ids of the validators, to which their bridge fees will be sent. Applies **only** for network with id `0`. If the bridge accepts Hedera Native Tokens, each member will need to have an association with the given token.                          |
| `bridge.networks[i].router_contract_address`           | ""      | The address of the Router contract on the EVM network. Ignored for network with id `0`.                                                                                                                                                                              |
| `bridge.networks[i].tokens.fungible[j]`                | ""      | The Address/HBAR/Token ID of the native fungible asset for the given network. Used as a key to for the following `bridge.networks[i].tokens.fungible[j].*` configuration fields below.                                                                               |
| `bridge.networks[i].tokens.fungible[j].fee_percentage` | ""      | The percentage which validators take for every bridge transfer. Applies **only** for assets from network with id `0`. Range is from 0 to 100.000 (multiplied by 1 000). Examples: 1% is 1 000, 1.234% = 1234, 0.15% = 150. Default 10% = 10 000                      |
| `bridge.networks[i].tokens.fungible[j].networks[k]`    | ""      | A key-value pair representing the id and wrapped asset to which the token `j` has a wrapped representation. Example: TokenID `0.0.2473688` (`j`) on Network `0` (`i`) has a wrapped version on `80001` (`k`), which is `0x95341E9cf3Bc3f69fEBfFC0E33E2B2EC14a6F969`. |
| `bridge.networks[i].tokens.nft[j]`                     | ""      | The Address/HBAR/Token ID of the native nft asset for the given network. Used as a key to for the following `bridge.networks[i].tokens.nft[j].*` configuration fields below.                                                                                         |
| `bridge.networks[i].tokens.nft[j].fee`                 | 0       | The HBAR fee (in tinybars), which validators take for every nft bridge transfer. Applies **only** for assets from network with id `0`. Default fee is 0, which is not be supported.                                                                                  |
| `bridge.networks[i].tokens.nft[j].networks[k]`         | ""      | A key-value pair representing the id and wrapped asset to which the token `j` has a wrapped representation. Example: TokenID `0.0.2473688` (`j`) on Network `0` (`i`) has a wrapped version on `80001` (`k`), which is `0x95341E9cf3Bc3f69fEBfFC0E33E2B2EC14a6F969`. |

### Run E2E Tests

```
go test ./e2e
```
`NB!` E2E tests are tightly coupled with the setup of [Three Validators Network](../examples/three-validators/README.md).