# Configuration
The application supports loading configuration from an `application.yml` file or via the environment.

Some configuration settings have appropriate defaults (e.g. database configuration) that can be left unchanged. 
On the other hand, important configuration settings like blockchain node endpoints, private keys,
account and topic ids, contract addresses and others have to be set, as they control the behaviour of the application.
Additionally, password properties have a default, but it is **strongly recommended passwords to be changed from the default**.

By default, the application loads a file named `application.yml` in each of the search paths (see below). The configuration loads
in the following order with the latter configuration overwriting the current configuration:

1. `./config/application.yml`
2. `./application.yml` (custom)
3. Environment variables, starting with `HEDERA_ETH_BRIDGE_` (e.g. `HEDERA_ETH_BRIDGE_CLIENT_NETWORK_TYPE=testnet`)

The following table lists the currently available properties, along with their default values.
Unless you need to set a non-default value, it is recommended to only populate overwritten properties in the custom `application.yml`.

Name                                                                | Default                                             | Description
------------------------------------------------------------------- | --------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------
`hedera.validator.db.host`                                          | 127.0.0.1                                           | The IP or hostname used to connect to the database.
`hedera.validator.db.name`                                          | hedera_validator                                    | The name of the database.
`hedera.validator.db.password`                                      | validator_pass                                      | The database password the processor uses to connect.
`hedera.validator.db.port`                                          | 5432                                                | The port used to connect to the database.
`hedera.validator.db.username`                                      | validator                                           | The username the processor uses to connect to the database.
`hedera.validator.port`                                             | 5200                                                | The port on which the application runs.
`hedera.eth.node_url`                                               | ""                                                  | The endpoint of the Ethereum node.
`hedera.eth.router_contract_address`                                | ""                                                  | The address of the Router contract.
`hedera.eth.bridge_waiting_blocks`                                  | 5                                                   | The number of blocks to wait before processing an ethereum event
`hedera.mirror_node.client_address`                                 | hcs.testnet.mirrornode.hedera.com:5600              | The HCS Mirror node endpoint. Depending on the Hedera network type, this will need to be changed.
`hedera.mirror_node.api_address`                                    | https://testnet.mirrornode.hedera.com/api/v1/       | The Hedera Rest API root endpoint. Depending on the Hedera network type, this will need to be changed.
`hedera.mirror_node.polling_interval`                               | 5                                                   | How often (in seconds) the application will poll the mirror node for new transactions.
`hedera.client.network_type`                                        | testnet                                             | Which Hedera network to use. Can be either `mainnet`, `previewnet`, `testnet`.
`hedera.client.operator.account_id`                                 | ""                                                  | The operator's Hedera account id.
`hedera.client.operator.private_key`                                | ""                                                  | The operator's Hedera private key.
`hedera.client.operator.eth_private_key`                            | ""                                                  | The operator's Ethereum private key.
`hedera.watcher.crypto-transfer.account.id`                         | ""                                                  | The Hedera account id to which the crypto transfer watcher will subscribe.
`hedera.watcher.crypto-transfer.account.max_retries`                | 10                                                  | The number of times the watcher will try restarting in case it failed.
`hedera.watcher.crypto-transfer.account.start_timestamp`            | ""                                                  | The timestamp from which the crypto transfer watcher will begin its subscription. Leave empty on the first run if you want to begin from `now`.
`hedera.watcher.consensus-message.topic.id`                         | ""                                                  | The Hedera topic id to which the consensus message watcher will subscribe.
`hedera.watcher.consensus-message.topic.max_retries`                | 10                                                  | The number of times the watcher will try restarting in case it failed.
`hedera.watcher.consensus-message.topic.start_timestamp`            | ""                                                  | The timestamp from which the consensus message watcher will begin its subscription. Leave empty on the first run if you want to begin from `now`.
`hedera.handler.crypto-transfer.topic_id`                           | ""                                                  | The Hedera topic id to which the crypto transfer handler will publish messages.
`hedera.handler.crypto-transfer.polling_interval`                   | 5                                                   | How often (in seconds) the crypto transfer handler will poll the status of a given transaction.
`hedera.handler.consensus-message.topic_id`                         | ""                                                  | The Hedera topic id to which the consensus message handler will publish messages.
`hedera.handler.consensus-message.send_deadline`                    | 300                                                 | The time (in seconds) between every execution window.
`hedera.rest_api_only`                                              | false                                               | The application will only expose REST API endpoints if this flag is true.
`hedera.log_level`                                                  | Info                                                | The log level of the validator. Possible values: `info`, `debug`, `trace` case insensitive. 


