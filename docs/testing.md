# Testing

## E2E Testing
Before you run E2E tests, you need to have a running application.

### Configuration
Configure properties, needed to run e2e tests in `e2e/config/application.yml`.
**Keep in mind that most of the configuration needs to be the same as the application's**.
It supports the following configurations:

Name                        | Description
-------------------------- | ----------------------------------
`hedera.bridge_account`    | The configured Bridge account. Validators listen for CryptoTranfers crediting that account
`hedera.topic_id`          | The configured Bridge Topic. Validators watch & submit signatures to that Topic
`hedera.sender.account`    | The account that will be sending the Hbars through the bridge
`hedera.sender.private_key`| The private key for the account that will be sending Hbars through the bridge
`hedera.ethereum.node_url` | The Ethereum Node that will be used for querying data
`hedera.ethereum.whbar_contract_address` | The ERC20 WHBAR contract on Ethereum
`hedera.ethereum.bridge_contract_address` | The Bridge contract on Ethereum
`hedera.validator_url` | The URL of the Validator node. Used for querying Metadata

### Run E2E Tests

```
go test ./e2e
```