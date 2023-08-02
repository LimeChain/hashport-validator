# Three Validators Bridge Network

## Overview

The goal of `Three Validators Bridge Network` is to showcase how validators configured for similar HCS topics and threshold accounts operate simultaneously.
The network consists of **three** validator nodes, which process incoming transfers and events, and a **read-only node**, used only to monitor the network and the bridge itself.

Example uses `Docker Compose`.

## Terminology

* HTS - Hedera token service.
* HCS - Hedera contract service.
* Native Token - In the sense of the Hashport Bridge, this is a token, that is deployed on the EVM or HTS and is the original version of that token. It is Native to the specific EVM. For example [Tether USD](https://etherscan.io/token/0xdac17f958d2ee523a2206206994597c13d831ec7) is Native to Ethereum but has been `bridged` to other EVMs. [WETH](https://etherscan.io/token/0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2) is also Native to Ethereum.
* Wrapped Token - In the sense of the Hashport Bridge, a wrapped token is a token that is deployed to EVM or HTS BY the bridge itself. For example [Theter USD](https://bscscan.com/token/0x55d398326f99059ff775485246999027b3197955) on BSC has been `bridged` by Binance and actually represents the value of the native [Ethereum Native Tether USD](https://etherscan.io/token/0xdac17f958d2ee523a2206206994597c13d831ec7).
* EVM - Ethereum Virtual Machine.

## Onboarding in depth guide

   To get a general idea for the whole hashport system it will be best to start with setting up the bridge.yml for the
   "Three Validators".

   ! Hedera scripts are found in the [Validator repo](https://github.com/LimeChain/hashport-validator)
   ! EVM scripts are found in [Contracts repo](https://github.com/LimeChain/hashport-contracts)
   
   1. go to [Hedera portal](https://portal.hedera.com/) and create a ED25519 Testnet ACCOUNT
   2. run the script to create the bridge topic (the network will be testnet and we will have 3 members):
      ```
      go run ./scripts/bridge/setup/cmd/setup.go \
         --privateKey=__ED25519_PRIVATE_KEY__ \
         --accountID=__ED25519_ACC__ \
         --adminKey=__ED25519_PUB_KEY__ \
         --network=testnet --members=3
      ```

      This script will generate
      - Alice, Bob and Carl accounts  
      - Bridge Account
      - Scheduled Tx Payer Account.
      ! Save it all.

   3. Create a Hedera Native token (save the output)
      ```
      go run ./scripts/token/native/create/cmd/create.go \
         --privateKey = __ED25519_PRIVATE_KEY__ \
         --accountID=__ED25519_ACC__ \
         --network=testnet \
         --memberPrKeys = __Alice__,__Bob__,__Carl__ \
         --bridgeID=__Bridge_ID__
      ```

   Now set up the EVM Routers. We will need at least 2 EVM networks, because this system can bridge directly from EVM to EVM.
   Possible bridging combinations include Hedera <-> EVM, EVM <-> EVM

   4. Setup EVM Router by running the [EVM deployment scripts](https://github.com/LimeChain/hashport-contracts) from the [hashport-contracts](https://github.com/LimeChain/hedera-eth-bridge-contracts/blob/main/README.md#scripts) repo. You will need 4 EVM Wallets: Owner, Alice, Bob and Carl accounts
   Deploy the router:
      ```
      npx hardhat deploy-router --help
         --fee-calculator-precision    The precision of fee calculations for native tokens (default: 100000)
         --governance-percentage       The percentage of how many of the total members are required to sign given message (default: 50)
         --governance-precision        The precision of division of required members signatures (default: 100)
         --members                     The addresses of the members 
         --members-admins              The addresses of the members' admins 
         --owner                       The owner of the to-be deployed router 

      npx hardhat deploy-router \
         --network __EVM_NETWORK__ \
         --owner __OWNER__ \
         --governance-percentage 50 \
         --governance-precision 100 \
         --fee-calculator-precision 100000 \
         --members __Alice__,__Bob__,__Carl__ \
         --members-admins __Alice__,__Bob__,__Carl__
      ```
      ! Save the router address

   5. To bridge native HBAR from Hedera we will need a wrapped version on the EVM network. The flow is Hedera HBAR --> Wrapped HBAR. To deploy a wrapped version of HBAR we will use:
      ```
      npx hardhat deploy-router-wrapped-token \
         --network __EVM_NETWORK__ \
         --router __EVM_ROUTER_ADDRESS__ \
         --source __EVM_CHAIN_ID__ \
         --native HBAR \
         --name "HBAR[__EVM_CHAIN_ID__]" \
         --symbol "HBAR[__EVM_CHAIN_ID__]" \
         --decimals 8
      ```

   6. To bridge a native token from the EVM to Hedera we will need to transfer the ownership of that token to the Router. This is a script that `Creates` a token and thenwe will `Transfers` ownership.
      ```
      npx hardhat deploy-token \
         --decimals 18 \
         --name "Some Native Token Name" \
         --symbol "SNTN" \
         --network __EVM_NETWORK__

      npx hardhat update-native-token \
         --fee-percentage 5 \
         --native-token __TOKEN_ADDRESS__ \
         --router __EVM_ROUTER_ADDRESS__ \
         --status true \
         --network __EVM_NETWORK__
      ```
 
   7. To pay fees for bridging from this EVM we will set a payment token.
      ```
      npx hardhat deploy-token \
         --decimals 6 \
         --name "USDC" \
         --symbol USDC \
         --network __EVM_NETWORK__

      npx hardhat set-payment-token \
         --network __EVM_NETWORK__ \
         --router __EVM_ROUTER_ADDRESS__ \
         --payment-token __TOKEN_ADDRESS__ \
         --status true
      ```

   8. Create a wrapped version of the Hedera token (from step 3) so we can bridge it. Native Hedera Token --> Wrapped EVM token
      ```
      npx hardhat deploy-router-wrapped-token \
         --network __EVM_NETWORK__ \
         --router __EVM_ROUTER_ADDRESS__ \
         --source __EVM_CHAIN_ID__ \
         --native __Token_ID__ \
         --name "__Name__" \
         --symbol "__SYMBOL__" \
         --decimals 8
      ```
   9. Create a wrapped version of HBAR token
      ```
      npx hardhat deploy-router-wrapped-token \
         --network __EVM_NETWORK__ \
         --router __EVM_ROUTER_ADDRESS__ \
         --source __EVM_CHAIN_ID__ \
         --native HBAR \
         --name "HBAR[__Name__]" \
         --symbol "HBAR[__SYMBOL__]" \
         --decimals 8
      ```

      ! [Hedera NAtive HBAR use 8 decimals, Tokens divide into 10 decimals pieces](https://docs.hedera.com/guides/docs/hedera-api/basic-types/tokenbalance). The `go run ./scripts/token/native/create/cmd/create.go` is set to create tokens with 8 decimals

   10. Use steps from 4 to 9 to deploy router to one more EVM network
   11. Create Wrapped versions for EVM to EVM bridging on both EVMs ( For Native Token on EVM we need corresponding Wrapped token on the Other EVM )
      ```
      npx hardhat deploy-router-wrapped-token \
         --network __EVM_NETWORK__ \
         --router __EVM_ROUTER_ADDRESS__ \
         --source __SOURCE_EVM_NETWORK_ID__ \
         --native __SOURCE_EVM_NATIVE_TOKEN_ADDRESS__ \
         --name "Wrapped __SOURCE_EVM_NATIVE_TOKEN_NAME__" \
         --symbol "W_ __SOURCE_EVM_NATIVE_TOKEN_SYMBOL__" \
         --decimals 18
      ```

   12. To enable bridging from EVM ---> Hedera. We will need to create a "Wrapped" versions of "EVM native token" on Hedera.
   Run wrapped-token-create.go to create custom wrapped token with a bridge account treasury and associate it with hedera (save the output)
      ```
      go run ./scripts/token/wrapped/create/cmd/create.go \
         --privateKey =__ED25519_PRIVATE_KEY__ \
         --accountID = __ED25519_ACC__ \
         --adminKey = __ED25519_PUB_KEY__ \
         --network = testnet \
         --memberPrKeys = __Alice__,__Bob__,__Carl__ \
         --bridgeID=__Bridge_ID__ \
         --generateSupplyKeysFromMemberPrKeys = true
      ```

   13. NFT setup. For now we can only bridge Hedera `Native` NFTs to other EVMs and those EVM wrapped versions back to the `Native` Hedera NFT.
      * Deploy NFT on Hedera
      ```
      go run ./scripts/token/native/nft/create/cmd/create.go \
         --privateKey=__ED25519_PRIVATE_KEY__ \
         --accountID=__ED25519_ACC__ \
         --network=testnet \
         --memberPrKeys=__Alice__,__Bob__,__Carl__ \
         --bridgeID=__Bridge_ID__
      ```

      * Mint the NFT on hedera
      ```
      go run scripts/token/native/nft/mint/main.go \
         -privateKey __ED25519_PRIVATE_KEY__ \
         -accountID __ED25519_ACC__ \
         -network testnet \
         -tokenID __ID_OF_DEPLOYED_NFT__ \
         -metadata SomeTestMetaData
      ```

      * Deploy the wrapped versions of the Hedera NFT to the EVMs
      ```
      npx hardhat deploy-wrapped-erc721-transfer-ownership \
         --network __EVM_NETWORK__ \
         --name "__WRAPPED_NFT_NAME__" \
         --router __EVM_ROUTER_ADDRESS__ \
         --symbol "__WRAPPED_NFT_SYMBOL__"
      ```

      * Set payment token the Hedera NFT to the EVMs
      ```
      npx hardhat set-payment-token \
         --network __EVM_NETWORK__ \
         --router __EVM_ROUTER_ADDRESS__ \
         --payment-token __TOKEN_ADDRESS__
      ```

!!! Make sure to have enough Tokens for paying fees and enough gas to use in each EVM Wallet.
!!! Make sure to associate all tokens with all hedera accounts

   ```
   go run ./scripts/token/associate/cmd/associate.go \
      --privateKey=__Wallet_PK__ \
      --accountID=__Wallet_ID__ \
      --network=testnet \
      --tokenID=__TOKEN_ID__
   ```

The structure of the three validators

```
examples
   |-- three-validators
   |   |-- alice
   |   |   |-- config
   |   |   |   |-- node.yml // has to be edited
   |   |-- bob
   |   |   |-- config
   |   |   |   |-- node.yml // has to be edited
   |   |-- carol
   |   |   |-- config
   |   |   |   |-- node.yml // has to be edited
   |   |-- dave
   |   |   |-- config
   |   |   |   |-- node.yml // has to be edited
   |   |-- docker-compose.yml
   |   |-- bridge.yml // has to be edited
```

## Validator Node YML setup

```YML
#Example Node
node:
  database:
    host: 127.0.0.1
    name: hedera_validator
    password: validator_pass
    port: 5432
    username: validator
  clients:
    evm:
      __EVM_CHAIN_ID_1__:
        block_confirmations: 5
        node_url:
          - __EVM_RPC__
        private_key: __WALLET_PK__ # This line will be different for Alice, Bob and Carl. Dave can use Alice config
      __EVM_CHAIN_ID_2__:
        block_confirmations: 5
        node_url: __EVM_RPC__
        private_key: __WALLET_PK__ # This line will be different for Alice, Bob and Carl. Dave can use Alice config
   hedera:
      operator:
        account_id: __NODE_ACC__ # from the scripts. This line will be different for Alice, Bob and Carl. Dave can use Alice config
        private_key: __NODE_PK__ # from the scripts. This line will be different for Alice, Bob and Carl. Dave can use Alice config
      network: testnet
   mirror_node:
      api_address: https://testnet.mirrornode.hedera.com/api/v1/
      client_address: hcs.testnet.mirrornode.hedera.com:5600
      polling_interval: 5
      query_default_limit: 25
      query_max_limit: 100
   coingecko:
      api_address: https://api.coingecko.com/api/v3/
   coin_market_cap:
      api_key: 81c414c6-376f-4cd5-8e91-a426417d6a5b # ask someone or get one from CoinMarketCap
      api_address: https://pro-api.coinmarketcap.com/v2/cryptocurrency/
  log_level: debug
  port: 5200
  recovery:
    start_timestamp:
    start_block:
  validator: true # This line will be set to "false" for DAVE, who is read-only validator node
```

## Bridge YML setup

```YML
bridge:
   use_local_config: true # set to true, because the validators will use this bridge yml config
   topic_id: __Bridge_TopicID__ # it was generated with `go run ./scripts/bridge/setup/cmd/setup.go`, find it after `Members Private keys array` and `Members Public keys array`
   networks:
    296: # Hedera
      name: "Hedera"
      bridge_account: __Bridge_Acc_ID__ 
      payer_account: __Payer_Acc_ID__ 
      members:
        - 0.0.48617120 __Alice_Acc_ID__ 
        - 0.0.48617123 __Bob_Acc_ID__ 
        - 0.0.48617124 __Carl_Acc_ID__ 
      tokens:
         fungible:
            "HBAR":
               fee_percentage: 10000 # 10.000%
               min_amount: 100
               min_fee_amount_in_usd: 0.2
               coin_gecko_id: "hedera-hashgraph"
               coin_market_cap_id: "4642"
               networks:
                  __EVM_CHAIN_ID_1__: "__Wrapped_HBAR_Address__"
                  __EVM_CHAIN_ID_2__: "__Wrapped_HBAR_Address__"
            "__HEDERA_NATIVE_TOKEN_ID__":
               fee_percentage: 10000
               min_amount: 100
               min_fee_amount_in_usd: 0.2
               coin_gecko_id: "tune-fm"
               coin_market_cap_id: "11420"
               networks:
                  __EVM_CHAIN_ID_1__: "__Wrapped_HBAR_Address__"
                  __EVM_CHAIN_ID_2__: "__Wrapped_HBAR_Address__"
         nft:
            "__HEDERA_NATIVE_NFT_ID__":
               fee: 20000
               networks:
                  __EVM_CHAIN_ID_1__: "__Wrapped_NFT_Address__"
                  __EVM_CHAIN_ID_2__: "__Wrapped_NFT_Address__"
   __EVM_CHAIN_ID_1__:
      name: "__EVM_CHAIN_NAME__"
      router_contract_address: "__EVM_ROUTER_ADDRESS__"
      tokens:
        fungible:
          "__EVM_Native_Token_address__":
            min_amount: 100
            min_fee_amount_in_usd: 0.2
            coin_gecko_id: "matic-network"
            coin_market_cap_id: "3890"
            networks:
               296: "__Wrapped_Token_ID__"
               __EVM_CHAIN_ID_2__: "__Wrapped_Token_Address__"
   __EVM_CHAIN_ID_2__:
      name: "__EVM_CHAIN_NAME__"
      router_contract_address: "__EVM_ROUTER_ADDRESS__"
      tokens:
        fungible:
          "__EVM_Native_Token_address__":
            min_amount: 100
            min_fee_amount_in_usd: 0.2
            coin_gecko_id: "matic-network"
            coin_market_cap_id: "3890"
            networks:
               296: "__Wrapped_Token_ID__"
               __EVM_CHAIN_ID_1__: "__Wrapped_Token_Address__"
```

## More documentation

1. Create a bridge configuration for Hedera using the [scripts](../../scripts/README.md).
2. Using the reference and the scripts for the Smart Contracts from [repository](https://github.com/LimeChain/hedera-eth-bridge-contracts/blob/main/README.md#scripts). Do the following: 
3. Set necessary [bridge](./bridge.yml) configuration for the nodes.
4. Set necessary configurations for [Alice](./alice/config), [Bob](./bob/config), [Carol](./carol/config) and [Dave](./dave/config).
