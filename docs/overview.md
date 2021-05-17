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

### Transferring HBAR/HTS from Hedera to the EVM chain

The transfer of assets from Hedera to the EVM chain is described in the following sequence diagram.
<p align="center">
  <img src="./assets/hedera-to-evm.png">
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
   `{hedera-tx-id}{router-address}{wrapped-token}{receiver}{amount}` using their EVM-compatible private key.
   The authorisation is then submitted to a topic in Hedera Consensus Service

5. **Waiting for Supermajority**
   Alice's UI or API waits for a supermajority of the signatures. She can either watch the topic messages stream or fetch the data directly from Validator nodes.

6. **Submitting the EVM Transaction**
   Once supermajority is reached, Alice submits the transaction to the EVM chain, claiming her wrapped asset. The signature contains the raw data signed in the message: `{hedera-tx-id}{router-address}{wrapped-token}{receiver}{amount}`

7. **Mint Operation**
   The smart contract verifies that no reply attack is being executed (by checking the `hedera-tx-id` and verifies the provided signatures against the raw data that was signed. If supermajority is reached, the `Router` contract `mints` the wrapped token to the `receiving` address.

### Transferring WHBAR/WHTS from the EVM chain to Hedera

The transfer of assets from the EVM chain to Hedera is described in the following sequence diagram.

<p align="center">
  <img src="./assets/evm-to-hedera.png">
</p>

#### Steps
1. **Initiating the Transfer**
   Alice wants to transfer her `WHBAR`/`WHTS` tokens from the EVM chain to Hedera. She opens any UI that integrates the Bridge and sends `burn` transaction to the `Router` contract. As parameter of the `burn` function, she specifies the Hedera account to receive the wrapped token.

2. **Burn Operation**
   The smart contract transfers the wrapped tokens from Alice's address and burns them. At the end of the process, a `Burn` event is emitted, containing the information about the burned token, the amount and the receiver.
3. **Picking up the Transfer**
   Validator nodes watch for `Burn` events and once such occurs, they prepare and submit `ScheduleCreate` operation that transfers the `service fee` amount from the `Bridge` account to the list of validators equally. Due to the nature of Scheduled Transactions, only one will be successfully executed, creating a scheduled Entity and all others will fail with `IDENTICAL_SCHEDULE_ALREADY_CREATED` error, and the transaction receipt will include the `ScheduleID` and the `TransactionID` of the first submitted transaction.
   All validators, except the one that successfully created the Transaction execute `ScheduleSign` and once `n out of m` validators execute the Sign operation, the transfer of the fees will be executed.
4. **Unlocking the Asset**
   Each Validator performs a `ScheduleCreate` operation that transfers `amount-serviceFee` `Hbar` to the receiving Hedera Account. All validators that got their `ScheduleCreate` rejected, submit an equivalent `ScheduleSign`. Once `n out of m` validators execute the Sign operation, the transfer is completed.