# Testing

## E2E Testing
Before you run E2E tests, you need to have a running application.

### Configuration
Configure properties, needed to run e2e tests in `e2e/config/application.yml`.
**Keep in mind that most of the configuration needs to be the same as the application's**.
It supports the following configurations:

Name                                              | Description
------------------------------------------------- | ----------------------------------
`hedera.bridge_account`                           | The configured Bridge account. Validators listen for CryptoTranfers crediting that account
`hedera.topic_id`                                 | The configured Bridge Topic. Validators watch & submit signatures to that Topic
`hedera.sender.account`                           | The account that will be sending the Hbars through the bridge
`hedera.sender.private_key`                       | The private key for the account that will be sending Hbars through the bridge
`hedera.ethereum.node_url`                        | The Ethereum Node that will be used for querying data
`hedera.ethereum.whbar_contract_address`          | The ERC20 WHBAR contract on Ethereum
`hedera.ethereum.bridge_contract_address`         | The Bridge contract on Ethereum
`hedera.validator_url`                            | The URL of the Validator node. Used for querying Metadata
`hedera.network_type`                             | Which Hedera network to use. Can be either `mainnet`, `previewnet`, `testnet`.
`hedera.tokens.whbar`                             | The whbar ID which will be used for the tests.
`hedera.tokens.wtoken`                            | The token ID which will be used for the token related tests.
`hedera.db_validation.host`                       | The IP or hostname used to connect to the database.
`hedera.db_validation.name`                       | The name of the database.
`hedera.db_validation.password`                   | The database password the processor uses to connect.
`hedera.db_validation.port`                       | The port used to connect to the database.
`hedera.db_validation.username`                   | The username the processor uses to connect to the database.

### Run E2E Tests

```
go test ./e2e
```