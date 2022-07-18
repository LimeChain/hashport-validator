# Overview

### Terminology

**Actors**
- **Users** - end-users that want to transfer HBAR or HTS tokens from Hedera to EVM-based chain or Wrapped HBAR and Wrapped Tokens from EVM-based chain to Hedera
- **Token Developers** - the developers of HTS tokens
- **Validators** - parties/entities that are running the Validator node. They are providing authorisation for the minting and burning of wrapped tokens on the EVM-based chain, as well as transferring wrapped tokens back to Hedera.
  
**Assets**
- **WHBAR** - ERC20 token issued and operated by Bridge validators. The token represents a "wrapped" HBAR on the EVM-based chain. In other words we can say that Hbar is the `native` asset and `WHBAR` is the `wrapped` asset.
- **WHTS** - ERC20 token issued and operated by Bridge validators. The token represents a "wrapped" HTS token on the EVM-based chain. In this case, `HTS` token is the `native` asset and `WHTS` token is the `wrapped` asset.

**Accounts**
- **Bridge Account** - Hedera threshold account (`n/m`), where `m` is the number of validators. Each validator has a Hedera compatible private key - 1 out of `m` that has `1/m` control over this threshold account. The funds transferred through the bridge (Hedera -> EVM) are sent to the Bridge account. The funds transferred back to Hedera (EVM -> Hedera) are sent from the Bridge account to the recipient's account.
- **Fee Account** - Hedera threshold account (`n/m`), where `m` is the number of validators. Each validator has a Hedera compatible private key - 1 out of `m` that has `1/m` control over this threshold account. The account is being used to pay for the transaction fees for transferring Assets from the bridge account to the recipient account.

### Governance

The setup on the EVM chain is the same - [Gnosis MultiSig](https://github.com/gnosis/safe-contracts) is to be used with the same `n/m` threshold configuration. Each validator has an EVM compatible private key - 1 out of `m` that as `1/m` control over the threshold account.
The Gnosis Multisig is configured as owner of:
- Wrapped tokens deployed on the EVM chain (f.e. WHBAR / WHTS tokens)
- The Router smart contract

#### Adding / Removing Members
Validators can add new members or remove members from the validator set. We expect validators to have an off-chain communication channel where validators can discuss the current setup, vote on removing validators or adding new ones.

Two transactions must be executed in order for the Validator set to change:
1. On Hedera, `Crypto Update` transaction that modifies the `n/m` threshold accounts (`Bridge` and `Fees`)
2. On the EVM chain, `updateMember` transaction that modifies the list of members in the `Router` contract (it may add or remove a member)

**Note**: Once a new validator is added, he can safely run a validator node with the correct credentials configured, and he will be authorising bridge transfers and accruing fees.

### Fees
The main incentive for becoming a Validator is the service fee paid by users. The fee is a percentage of the transferred amount, paid on the native asset.
For example transferring 100 HBAR from Hedera to the EVM chain (WHBAR) is going to have 1% service fee (1 HBAR) transferred to the Bridge Validators.

*Note: The Service fee is configurable property and determined by the validators*

## Hedera Fungible Native Assets

### Hedera to EVM

The transfer of assets from Hedera to the EVM chain is described in the following sequence diagram.
<p align="center">
  <img src="./assets/hedera-to-evm(fungible%20hedera%20native).png">
</p>

#### Steps
1. **Initiating the transfer**
   Alice wants to transfer `HBAR`/`HTS token` from Hedera to the EVM chain. She opens any UI that integrates the Bridge and sends the asset to the `Bridge` Account. The memo of the transfer contains the `evm-address`, which is going to be the receiver of the wrapped asset on the other chain.
2. **Picking up the Transfer**
   The Bridge validator nodes listen for new incoming transfers to the `Bridge` Account. Once they pick up the new transaction, they verify the `state proof` and validate that the `memo` contains a valid EVM address configured as receiver of the wrapped asset.
3. **Paying out fees**

   **3.1** Each of the Validators create a Schedule Create transaction transferring the `service fee` amount from the `Bridge` account to the list of validators equally *(f.e. if the service fee is `7 HBAR` and there are `7` validators, the Schedule Create Transfer will contain Transfer list crediting `1 HBAR` to each of the validators.)*

   **3.2** Due to the nature of Scheduled Transactions, only one will be successfully executed, creating a scheduled Entity and all others will fail with `IDENTICAL_SCHEDULE_ALREADY_CREATED` error, and the transaction receipt will include the `ScheduleID` of the first submitted transaction.
All validators, except the one that successfully created the Transaction execute `ScheduleSign` and once `n out of m` validators execute the Sign operation, the transfer of the fees will be executed.

4. **Providing Authorisation Signature**
   Each of the Validators sign the following authorisation message:
   `{source-chain-id}{target-chain-id}{hedera-tx-id}{wrapped-token}{receiver}{amount}` using their EVM-compatible private key.
   The authorisation is then submitted to a topic in Hedera Consensus Service

5. **Waiting for Supermajority**
   Alice's UI or API waits for a supermajority of the signatures. She can either watch the topic messages stream or fetch the data directly from Validator nodes.

6. **Submitting the EVM Transaction**
   Once supermajority is reached, Alice submits the transaction to the EVM chain, claiming her wrapped asset. The transaction contains the raw data signed in the message: `{source-chain-id}{target-chain-id}{hedera-tx-id}{wrapped-token}{receiver}{amount}`

7. **Mint Operation**
   The smart contract verifies that no reply attack is being executed (by checking the `hedera-tx-id` and verifies the provided signatures against the raw data that was signed. If supermajority is reached, the `Router` contract `mints` the wrapped token to the `receiving` address.

### EVM to Hedera

The transfer of assets from the EVM chain to Hedera is described in the following sequence diagram.

<p align="center">
  <img src="./assets/evm-to-hedera(fungible%20hedera%20native).png">
</p>

#### Steps
1. **Initiating the Transfer**
   Alice wants to transfer her `WHBAR`/`WHTS` tokens from the EVM chain to Hedera. She opens any UI that integrates the Bridge and sends `burn` transaction to the `Router` contract. As parameter of the `burn` function, she specifies the Hedera account to receive the native fungible token.

2. **Burn Operation**
   The smart contract transfers the wrapped tokens from Alice's address and burns them. At the end of the process, a `Burn` event is emitted, containing the information about the burned token, the amount and the receiver.
3. **Picking up the Transfer**
   Validator nodes watch for `Burn` events and once such occurs, they prepare and submit `ScheduleCreate` operation that transfers the `service fee` amount from the `Bridge` account to the list of validators equally. Due to the nature of Scheduled Transactions, only one will be successfully executed, creating a scheduled Entity and all others will fail with `IDENTICAL_SCHEDULE_ALREADY_CREATED` error, and the transaction receipt will include the `ScheduleID` and the `TransactionID` of the first submitted transaction.
   All validators, except the one that successfully created the Transaction execute `ScheduleSign` and once `n out of m` validators execute the Sign operation, the transfer of the fees will be executed.
4. **Unlocking the Asset**
   Each Validator performs a `ScheduleCreate` operation that transfers `amount-serviceFee` `Hbar` to the receiving Hedera Account. All validators that got their `ScheduleCreate` rejected, submit an equivalent `ScheduleSign`. Once `n out of m` validators execute the Sign operation, the transfer is completed.


## EVM Fungible Native Assets
In order for an EVM native asset to be bridged to Hedera and mapped ot HTS token, the Governance mechanism must:
1. Deploy the corresponding HTS token
2. Map the EVM token to the HTS token
3. Whitelist the EVM asset in the contract and specify the service fee

The HTS token mapped to EVM token has the following configuration:
- The name and symbol of the Token is the same as the EVM one
- Decimals of the token are set to `8` due to HTS <> EVM compatibility issues
- The `bridge account` is used as a `treasury` for the Token. This enforces the "owner" of the token to be the shared Hedera account that is governed by the Validator set. 
- The `supplyKey` for the Token is a `ThresholdKey` equivalent to the `bridge account`, meaning that in order for tokens to be minted or burned, `n/m` validators must sign the `mint/burn` transactions.

Once the steps above are performed, a given EVM asset can be transferred through the bridge in both directions.

### EVM to Hedera
The following sequence diagram demonstrates the process of transferring ERC20 Token from EVM chain to Hedera:

<p align="center">
  <img src="./assets/evm-to-hedera(fungible%20evm%20native).png">
</p>

#### Steps

1. **Lock** - Performed by the User
Alice calls the `lock` function of the `Router` contract specifying the `address` of the Token, the `amount` she wants to bridge and the recipient `Hedera Account` that will receive the tokens.
The contract verifies that the specified token is supported, transfers the Token from `Alice`s account, charges a service fee (% of the token amount), distributes the fee to all validators of the Bridge and emits a `lock`-ing event with all of the required metadata.
2. **Event Monitoring** - Ongoing process performed by Validator nodes
Validator nodes are monitoring the `router` contract for `lock` events. Once such event is emitted, each validator computes the corresponding `Hedera HTS` token `ID` of the bridged asset. 
3. **Minting the Tokens** - Performed by Validators
Validators create `Scheduled Mint` transactions. Once the required `n/m` keys have executed either `ScheduleCreate` or `ScheduleSign` operation, the specified in the `EVM` transaction tokens, are minted to the `treasury`. (Note: `TokenMint` operation can mint tokens only to the treasury).
4. **Transferring the Tokens** - Performed by Validators
Validators create `ScheduleTransfer` transactions that transfer the newly minted tokens from the `Treasury` to Alice's `Hedera Account`.

### Hedera to EVM
The following sequence diagram demonstrates the process of transferring HTS Tokens mapped to ERC20 Tokens from Hedera to the source EVM chain:

<p align="center">
  <img src="./assets/hedera-to-evm(fungible%20evm%20native).png">
</p>

#### Steps
1. **Initiate the transfer** - Performed by the User
Alice executes a `CryptoTransfer` operation sending the mapped HTS tokens (that she wants to transfer back to the EVM chain) to the corresponding `treasury` account of the HTS tokens. In the `memo` of the transaction, she encodes the following information: `{chainId}-{receiving-address}`, where `chainId` is the [chain ID used in EVM based chains](https://docs.soliditylang.org/en/latest/units-and-global-variables.html#block-and-transaction-properties) and `{receiving-address}` is the EVM address that will receive the EVM native tokens.
2. **Transfers Monitoring** - Ongoing process performed by Validators
Validator nodes are monitoring new incoming transfers towards the configured `treasury` (reused for all Bridge supported tokens). Once such transfer is picked up, Validators are executing state proof verification and proceed with the bridging.
3. **Burning the tokens** - Performed by Validators
Validators creates scheduled `TokenBurn` operation that removes the `amount` sent to the treasury from the treasury account. Once `n/m` keys (validators) have executed either `ScheduleCreate` or `ScheduleSign` operation, the specified amount of tokens gets burned and the total supply of the `HTS` token is reduced.
4. **Providing authorisation signature** - Performed by Validators
Each of the Validators sign the following authorisation message:
   `{source-chain-id}{target-chain-id}{hedera-tx-id}{wrapped-token}{receiver}{amount}` using their EVM-compatible private key. The signature is then submitted to a topic in Hedera Consensus Service.
5. **Waiting for Supermajority**
   Alice waits for a supermajority of the signatures. She can either watch the topic messages stream or fetch the data directly from Validator nodes.
6. **Unlocking the EVM native tokens** - Performed by the User
Once supermajority is reached, Alice submits `unlock` transaction to the EVM chain. The transaction contains the raw data signed in the authorisation signatures, as-well as the signatures. The smart contract verifies the authenticity of the signatures, charges `service` fee and transfers the requested token to the specified `recipient` address.

## Hedera Non-Fungible Native Assets

### Hedera to EVM

The transfer of NFT assets from Hedera to the EVM chain is described in the following sequence diagram.
<p align="center">
  <img src="./assets/hedera-to-evm%28nft%20hedera%20native%29%20using%20crypto%20allowances.png">
</p>

#### Steps
1. **Initiating the transfer**
   
   **1.1** Alice wants to transfer `HTS NFT` from Hedera to the EVM chain. She opens any UI that integrates the Bridge.
   
   **1.2** Alice sends `CryptoApproveAllowance` transaction for the given NFT, specifying the spender to be `Payer Account` of the bridge configuration.

   **1.3** Alice sends the flat fee and all the necessary Royalty (fallback fee) for the NFT asset to the `Bridge` Account. The memo of the transfer contains the other `chain-Id`, the `evm-address` which is going to be the receiver of the wrapped ERC-721 (NFT) asset and the Hedera NFT ID `serial@token-id`.

2. **Picking up the Transfer**
   The Bridge validator nodes listen for new incoming transfers to the `Bridge` Account. Once they pick up the new transaction, they validate that the `memo` contains a valid EVM address configured as receiver of the wrapped asset and the NFT ID of the Native asset is valid.
3. **Transferring the NFT to the Bridge Account**
   
   **3.1** Each of the Validators create a Schedule Create transaction, which transfers the `NFT` to the `Bridge` account.

   **3.2** Due to the nature of Scheduled Transactions, only one will be successfully executed, creating a scheduled Entity and all others will fail with `IDENTICAL_SCHEDULE_ALREADY_CREATED` error, and the transaction receipt will include the `ScheduleID` of the first submitted transaction.

4. **Paying out fees**

   **4.1** Each of the Validators create a Schedule Create transaction transferring the `flat NFT transfer fee` amount from the `Bridge` account to the list of validators equally.

   **4.2** Due to the nature of Scheduled Transactions, only one will be successfully executed, creating a scheduled Entity and all others will fail with `IDENTICAL_SCHEDULE_ALREADY_CREATED` error, and the transaction receipt will include the `ScheduleID` of the first submitted transaction.
   All validators, except the one that successfully created the Transaction execute `ScheduleSign` and once `n out of m` validators execute the Sign operation, the transfer of the fees will be executed.

5. **Providing Authorisation Signature**
   Each of the Validators sign the following authorisation message:
   `{source-chain-id}{target-chain-id}{hedera-tx-id}{wrapped-token}{token-id}{metadata}{receiver}` using their EVM-compatible private key.
   The authorisation is then submitted to a topic in Hedera Consensus Service

6. **Waiting for Supermajority**
   Alice's UI or API waits for a supermajority of the signatures. She can either watch the topic messages stream or fetch the data directly from Validator nodes.

7. **Submitting the EVM Transaction**
   Once supermajority is reached, Alice submits the transaction to the EVM chain, claiming her wrapped asset. The transaction contains the raw data signed in the message: `{source-chain-id}{target-chain-id}{hedera-tx-id}{wrapped-token}{token-id}{metadata}{receiver}`

8. **Mint Operation**
   The smart contract verifies that no reply attack is being executed by verifying the provided signatures against the raw data that was signed. If supermajority is reached, the `Router` contract `mints` the wrapped ERC-721 (NFT) to the `receiving` address.

### EVM to Hedera

The transfer of ERC-721 (NFT) assets from the EVM chain to Hedera is described in the following sequence diagram.

<p align="center">
  <img src="./assets/evm-to-hedera(nft%20hedera%20native).png">
</p>

#### Steps
1. Alice wants to transfer her **wrapped** `NFT` from the EVM chain back to Hedera. Before initiating the actual transfer, Alice needs to:

   1.1 **Approve flat fee**
      Alice submits an approve ERC-20 transaction, which approves the flat fee amount to the `Router` contract.

   1.2 **Approve ERC-721 (NFT)**
      Alice submits an approve ERC-721 (NFT) transaction, which approves the NFT to be burnt by the `Router` contract.

2. **Initiating the Transfer**
   Alice sends `burn` transaction to the `Router` contract. As parameter of the `burn` function, she specifies the NFT and the Hedera account to receive the Hedera native NFT.
3. **Burn Operation**
   The smart contract transfers the flat fee from Alice's address, and burns the NFT. At the end of the process, a `Burn` event is emitted, containing the information about the NFT and the receiver.
4. **Picking up the Transfer**
   Validator nodes watch for `Burn` events and once such occurs, they prepare and submit `ScheduleCreate` operation of `CryptoApproveAllowance` that specifies the `spender` of the `NFT` as the receiver. Due to the nature of Scheduled Transactions, only one will be successfully executed, creating a scheduled Entity and all others will fail with `IDENTICAL_SCHEDULE_ALREADY_CREATED` error, and the transaction receipt will include the `ScheduleID` and the `TransactionID` of the first submitted transaction.
   All validators, except the one that successfully created the Transaction execute `ScheduleSign` and once `n out of m` validators execute the Sign operation, the transfer of the fees will be executed.
5. **Unlocking the Asset**
   After the NFT has been approved for the user, he can then submit a transfer transaction, taking it back in his account.