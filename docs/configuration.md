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
`hedera.eth.block_confirmations`                                    | 5                                                   | The number of block confirmations to wait for before processing an ethereum event
`hedera.mirror_node.client_address`                                 | hcs.testnet.mirrornode.hedera.com:5600              | The HCS Mirror node endpoint. Depending on the Hedera network type, this will need to be changed.
`hedera.mirror_node.api_address`                                    | https://testnet.mirrornode.hedera.com/api/v1/       | The Hedera Rest API root endpoint. Depending on the Hedera network type, this will need to be changed.
`hedera.mirror_node.polling_interval`                               |                                                     | How often (in seconds) the application will poll the mirror node for new transactions.
`hedera.mirror_node.account_id`                                     |                                                     | The account id that the validators use to monitor for incoming transfers.
`hedera.mirror_node.topic_id`                                       |                                                     | The topic id that the validators use to monitor for incoming hedera consensus messages.
`hedera.client.network_type`                                        | testnet                                             | Which Hedera network to use. Can be either `mainnet`, `previewnet`, `testnet`.
`hedera.client.operator.account_id`                                 | ""                                                  | The operator's Hedera account id.
`hedera.client.operator.private_key`                                | ""                                                  | The operator's Hedera private key.
`hedera.client.operator.eth_private_key`                            | ""                                                  | The operator's Ethereum private key.
`hedera.watcher.recovery_timestamp`                                 |                                                     | The timestamp from which the crypto transfer watcher will begin its recovery. Leave empty on the first run if you want to begin from `now`.
`hedera.watcher.max_retries`                                        | 10                                                  | The number of times the watcher will try restarting in case it failed.
`hedera.handler.send_deadline`                                      | 300                                                 | The time (in seconds) between every execution window.
`hedera.rest_api_only`                                              | false                                               | The application will only expose REST API endpoints if this flag is true.
`hedera.log_level`                                                  | info                                                | The log level of the validator. Possible values: `info`, `debug`, `trace` case insensitive. 


