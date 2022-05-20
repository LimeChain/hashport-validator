# Timestamp & Originator columns migrations

This script populates the `timestamp` and `originator` columns in the `transfers` table.
The script will sequentially run through all dbs listed in the configuration file.



For hedera TXs, it fetches the tx from the mirror node, and sets the consensus timestamp.
For EVM TXs, we get the block timestamp.


For each EVM network, you will have to supply API providers. This is done in the `evm` key-value section of the configuration.

For the mirror node communication you will need to supply an `api_address`, `client_address` corresponding to the Hedera network you want to use (testnet, mainnet).