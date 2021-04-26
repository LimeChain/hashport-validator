# Testing

## E2E Testing

Before you run E2E tests, you need to have a running application.

### Configuration

Configure properties, needed to run e2e tests in `e2e/config/application.yml`.
**Keep in mind that most of the configuration needs to be the same as the application's**.
It supports the following configurations:

Name                                      | Description                                                                                                                                                                                                       |
----------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
`hedera.bridge_account`                   | The configured Bridge account. Validators listen for CRYPTOTRANSFER transactions, crediting that account.
`hedera.dbs[].host`                       | The IP or hostname used to connect to the database.
`hedera.dbs[].name`                       | The name of the database.
`hedera.dbs[].password`                   | The database password the processor uses to connect.
`hedera.dbs[].port`                       | The port used to connect to the database.
`hedera.dbs[].username`                   | The username the processor uses to connect to the database.
`hedera.fee_percentage`                   | The percentage which validators take for every bridge transfer. Range is from 0 to 100.000 (multiplied by 1 000). Examples: 1% is 1 000, 1.234% = 1234, 0.15% = 150. Default 10% = 10 000.
`hedera.members`                          | The Hedera account ids of the validators, to which their bridge fees will be sent (if Bridge accepts Hedera Tokens, associations with these tokens will be required). Used to assert balances after transactions.
`hedera.mirror_node.api_address`          | The Hedera Rest API root endpoint. Depending on the Hedera network type, this will need to be changed.
`hedera.mirror_node.client_address`       | The HCS Mirror node endpoint. Depending on the Hedera network type, this will need to be changed.
`hedera.mirror_node.max_retries`          | The maximum number of retries that the mirror node has to continue monitoring after a failure, before stopping completely.
`hedera.mirror_node.polling_interval`     | How often (in seconds) the application will poll the mirror node for new transactions.
`hedera.network_type`                     | Which Hedera network to use. Can be either `mainnet`, `previewnet`, `testnet`.
`hedera.sender.account`                   | The account that will be sending assets through the bridge.
`hedera.sender.private_key`               | The private key for the account that will be sending assets through the bridge.
`hedera.topic_id`                         | The configured Bridge Topic. Validators watch & submit signatures to that Topic.
`ethereum.block_confirmations`            | The number of block confirmations to wait for before processing an ethereum event.
`ethereum.node_url`                       | The Ethereum Node that will be used for querying data.
`ethereum.private_key`                    | The private key for the account which executes mint/burn operations. The derived address of the key serves as an `HBAR->Ethereum` memo (receiver).
`ethereum.router_contract_address`        | The address of the Router Contract.
`tokens.whbar`                            | The native asset, which represents HBAR. Used as a bridged asset between the two networks.
`tokens.wtoken`                           | The native asset, which represents a Token on Hedera. Used as a bridged asset between the two networks.
`validator_url`                           | The URL of the Validator node. Used for querying Metadata.

### Run E2E Tests

```
go test ./e2e
```
