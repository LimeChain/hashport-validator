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
3. Environment variables, starting with `VALIDATOR_` (e.g. `VALIDATOR_CLIENTS_HEDERA_NETWORK_TYPE=testnet`)

The following table lists the currently available properties, along with their default values.
Unless you need to set a non-default value, it is recommended to only populate overwritten properties in the custom `application.yml`.

Name                                                                | Default                                             | Description
------------------------------------------------------------------- | --------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------
`validator.database.host`                                           | 127.0.0.1                                           | The IP or hostname used to connect to the database.
`validator.database.name`                                           | hedera_validator                                    | The name of the database.
`validator.database.password`                                       | validator_pass                                      | The database password the processor uses to connect.
`validator.database.port`                                           | 5432                                                | The port used to connect to the database.
`validator.database.username`                                       | validator                                           | The username the processor uses to connect to the database.
`validator.clients.ethereum.block_confirmations`                    | 5                                                   | The number of block confirmations to wait for before processing an ethereum event
`validator.clients.ethereum.node_url`                               | ""                                                  | The endpoint of the Ethereum node.
`validator.clients.ethereum.private_key`                            | ""                                                  | The operator's Ethereum private key.
`validator.clients.ethereum.router_contract_address`                | ""                                                  | The address of the Router contract.
`validator.clients.hedera.operator.account_id`                      | ""                                                  | The operator's Hedera account id.
`validator.clients.hedera.operator.private_key`                     | ""                                                  | The operator's Hedera private key.
`validator.clients.hedera.bridge_account`                           | ""                                                  | The account id validators use to monitor for incoming transfers. Also, serves as a distributor for Hedera transfers (validator fees and bridged amounts).
`validator.clients.hedera.network_type`                             | testnet                                             | Which Hedera network to use. Can be either `mainnet`, `previewnet`, `testnet`.
`validator.clients.hedera.payer_account`                            | ""                                                  | The account id paying for Hedera transfers fees.
`validator.clients.hedera.topic_id`                                 | ""                                                  | The topic id that the validators use to monitor for incoming hedera consensus messages.
`validator.clients.mirror_node.api_address`                         | https://testnet.mirrornode.hedera.com/api/v1/       | The Hedera Rest API root endpoint. Depending on the Hedera network type, this will need to be changed.
`validator.clients.mirror_node.client_address`                      | hcs.testnet.mirrornode.hedera.com:5600              | The HCS Mirror node endpoint. Depending on the Hedera network type, this will need to be changed.
`validator.clients.mirror_node.max_retries`                         | 10                                                  | The maximum number of retries that the mirror node has to continue monitoring after a failure, before stopping completely.
`validator.clients.mirror_node.polling_interval`                    | 5                                                   | How often (in seconds) the application will poll the mirror node for new transactions.
`validator.log_level`                                               | info                                                | The log level of the validator. Possible values: `info`, `debug`, `trace` case insensitive.
`validator.port`                                                    | 5200                                                | The port on which the application runs.
`validator.recovery.start_timestamp`                                | ""                                                  | The timestamp from which the crypto transfer watcher will begin its recovery. Leave empty on the first run if you want to begin from `now`.
`validator.rest_api_only`                                           | false                                               | The application will only expose REST API endpoints if this flag is true.
