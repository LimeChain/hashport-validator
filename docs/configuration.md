# Configuration
The application supports loading configuration from an `application.yml` file or via the environment.

Some configuration settings have appropriate defaults (e.g. database configuration) that can be left unchanged. 
On the other hand, important configuration settings like blockchain node endpoints, private keys,
account and topic ids, contract addresses and others have to be set, as they control the behaviour of the application.
Additionally, password properties have a default, but it is **strongly recommended passwords to be changed from the default**.

By default, it loads a file named `application.yml` in each of the search paths (see below). The configuration loads
in the following order with the latter configuration overwriting the current configuration:

1. `./config/application.yml`
2. `./application.yml`
3. Environment variables, starting with `HEDERA_ETH_BRIDGE_` (e.g. `HEDERA_ETH_BRIDGE_CLIENT_NETWORK_TYPE=testnet`)

The following table lists the currently available properties, along with their default values.
Unless you need to set a non-default value, it is recommended to only populate overwritten properties in the custom `application.yml`.

Name                                                    | Default                                             | Description
------------------------------------------------------- | --------------------------------------------------- | ----------------------------------------------------------------------------------------------
`hedera.validator.db.host`                              | 127.0.0.1                                           | The IP or hostname used to connect to the database
`hedera.validator.db.name`                              | hedera_validator                                    | The name of the database
`hedera.validator.db.password`                          | validator_pass                                      | The database password the processor uses to connect.
`hedera.validator.db.port`                              | 5432                                                | The port used to connect to the database
`hedera.validator.db.username`                          | validator                                           | The username the processor uses to connect to the database
`hedera.validator.port`                                 | 5200                                                |
`hedera.eth.node_url`                                   |                                                     |
`hedera.eth.bridge_contract_address`                    |                                                     |
`hedera.mirror_node.client_address`                     | hcs.testnet.mirrornode.hedera.com:5600              |
`hedera.mirror_node.api_address`                        | https://testnet.mirrornode.hedera.com/api/v1/       |
`hedera.mirror_node.polling_interval`                   | 5                                                   |
`hedera.client.network_type`                            | testnet                                             | Which Hedera network to use. Can be either `mainnet`, `previewnet`, `testnet`
`hedera.client.operator.account_id`                     |                                                     |
`hedera.client.operator.private_key`                    |                                                     |
`hedera.client.operator.eth_private_key`                |                                                     |
`hedera.watcher.crypto-transfer.accounts`               |                                                     |
`hedera.watcher.consensus-message.topics`               |                                                     |
`hedera.handler.crypto-transfer.topic_id`               |                                                     |
`hedera.handler.crypto-transfer.polling_interval`       |                                                     |
`hedera.handler.consensus-message.topic_id`             |                                                     |
`hedera.handler.consensus-message.addresses`            |                                                     |
`hedera.handler.consensus-message.send_deadline`        | 300                                                 |


