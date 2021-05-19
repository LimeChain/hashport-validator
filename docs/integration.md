# Integration with Hedera <-> EVM-chain bridge
The Bridge enables users to transfer HBAR or HTS tokens from Hedera to EVM-based chain or Wrapped HBAR and Wrapped Tokens from EVM-based chain to Hedera. It is operated by registered validators that provide signatures for every requested transfer. The transfer is processed when the majority of validators verify the transfer (supermajority).

## Transfers from Hedera to EVM-chain

This functionality allows users to transfer HBAR or any HTS token supported by the bridge and receive a wrapped version of the asset on the EVM chain.

### Step 1. Deposit Transaction

In order to initiate transfer, the user needs to submit a deposit transaction to the Hedera Bridge Account. More information on the account can be found in the [overview](./overview.md) document.

The transfer **must** specify the receiving address in the `memo` field.

Example:

In order to transfer **100 HBAR-s** to address _0x700d8a76b37f672a06ab89fe1ec95acfba799f1c_, the user needs to create a `CryptoTransfer` of **100 HBAR-s** to the bridge account and **add the receiving EVM address as MEMO to the transfer**.

>Transfer amount: 100 HBAR
>Memo: 0x700d8a76b37f672a06ab89fe1ec95acfba799f1c

The corresponding `TransactionID` is a unique identifier for the Bridge transfer operation and can be used to query the status of the Bridge transfer.

### Step 2. Waiting for Signatures

Bridge operators (validators) submit their signatures to an `HCS` topic, however for the ease of use, clients can query the signatures for a given transfer directly from the Validator's API:

    GET {validator_url}:{port}/api/v1/transfers/{transaction_id}

Where `transaction_id` is the Hedera `TransactionID` of the `CryptoTransfer` sending the asset to the Bridge account.

The response is in JSON format and contains the following data:

```json
{
  "recipient": "0x700d8a76b37f672a06ab89fe1ec95acfba799f1c",
  "routerAddress": "0x",
  "amount": "100",
  "nativeAsset": "",
  "wrappedAsset": "",
  "signatures": [
  ],
  "majority": false
}
```
Property | Description
---------- | ----------
**Recipient** | EVM address of the receiver
**RouterAddress** | Address of the router contract
**Amount** | Transfer amount
**NativeAsset** | Alias for the transferred asset
**WrappedAsset** | Alias for the wrapped asset
**Signatures** | Array of all provided signatures by the validators up until this moment
**Majority** | True if supermajority is reached and the wrapped token may be claimed

### Step 3. Claiming Wrapped Asset
Once supermajority is reached the users can claim _wrapped version_ of the asset. In order to do that, the user must sign and submit a **mint transaction** to the Bridge Router Contract.

The mint operation can be constructed using the following arguments:

	mint(transactionId: bytes , wrappedAsset: address, receiver: address, amount: UInt256, signatures: bytes[] )

Argument | Description
---------- | ----------
**transactionId** | The Hedera `TransactionID` of the Deposit transaction.  TODO add info on how to transform TXID to bytes
**wrappedAsset** | The corresponding `wrappedAsset` to claim. Must be the same as `wrappedAsset` returned from the Validator API query.
**receiver** | The address receiving the tokens. Must be the same as the one specified in the `memo`
**amount** | The amount to be minted. Keep in mind that this amount is `amount=original-serviceFee`. TODO verify it
**signatures** | The array of signatures provided by the Validator API

### Service Fee

The main incentive for the Validators is the `service fee` charged on every transfer. The fee is a percentage of the transferred amount, paid on the native asset. The Service fee is configurable property and determined by the validators.
Fees are paid out from the Bridge account.


## From EVM chain to Hedera
This functionality allows the user to transfer Wrapped HBAR or any supported by the bridge Wrapped Tokens from EVM-based chain to Hedera.

### Query supported tokens

In order to get all the supported wrapped tokens by the bridge the user must do two things:

1. Get the wrapped tokens count from the Router contract by calling the function `wrappedAssetsCount()`.
2. Call the Router contract function `wrappedAssetAt(uint256  index)` for every value between _0 ... wrappedAssetsCount-1_. Each time the function will return the **address** of the ERC20 wrapped token contract.

In order to map the wrapped assets to their native representation in Hedera, one can query the following mapping:

`mapping(address => bytes) public wrappedToNative;`

where  `address`  represents the ERC20 address of the wrapped asset and  `bytes`  represent the HTS Entity ID or simply  `HBAR`  (in the case for HBAR-s).

### Step 1. Burn the Wrapped asset

Transfers from EVM chain to Hedera is achieved by submitting a **burn operation** to the Router Contract.
There are two supported contract functions by which this can be done:

#### Option 1 - Approve + Burn

The straightforward way for burning the wrapped assets would be by executing 2 transactions:
1. ERC20 Approve
2. Router Burn 

The burn transaction has the following format:
- `burn(amount: uint256, receiver: bytes, wrappedAsset: address)`

Argument | Description
---------- | ----------
**amount** | The amount of wrapped tokens to be burned and transferred in their native version on Hedera
**receiver** | The Hedera Account to receive the native representation of the wrapped asset
**wrappedAsset** | The corresponding wrapped asset to burn

>Note: The receiver [AccountId](https://hashgraph.github.io/hedera-sdk-java/index.html?com/hedera/hashgraph/sdk/account/AccountId.html) must be serialized by Hedera SDK as such:
>`accountId._toProto().serializeBinary()`, before passing it as a argument to the _burn_ function.

#### Option 2 - Sign Permit + Burn

Using the Permit design, we are able to initiate the transfer of the assets in one transaction. Instead of the user executing a separate ERC20 `approve` TX, he must sign a `permit` message that is verified by the ERC20 contract in order to authorise the Router contract to spend the user's funds on his behalf.


Here is a example on how to create the necessary signature for permit operation:

```typescript
async function createPermit(  
        owner,  
        spenderAddress,  
        amount,  
        deadline,  
        tokenContract  
    ) {  
        const Permit = [  
            { name: "owner", type: "address" },  
            { name: "spender", type: "address" },  
            { name: "value", type: "uint256" },  
            { name: "nonce", type: "uint256" },  
            { name: "deadline", type: "uint256" },  
        ];        
        const domain = {  
            name: await tokenContract.name(),  
            version: "1",  
            chainId: "31337",  
            verifyingContract: tokenContract.address,  
        };        
        const message = {  
            owner: owner.address,  
            spender: spenderAddress,  
            value: amount,  
            nonce: await tokenContract.nonces(owner.address),  
            deadline: deadline,  
        };        
        const result = await owner._signTypedData(domain, { Permit }, message);  
        return {  
            r: result.slice(0, 66),  
            s: "0x" + result.slice(66, 130),  
            v: parseInt(result.slice(130, 132), 16),  
        };  
    }
```
> Note: **deadline** is the timestamp, after which the _permit_ for _burn operation_ will not be active. The user can set it however he likes.

> More information about the _permit_ operation can be found here: [EIP 2612](https://eips.ethereum.org/EIPS/eip-2612)

Once the user signs the `permit`, the `burnWithPermit` transaction can be executed. The signature (`v`, `r` and `s`) along with the `deadline` are send as part of the burn transaction: 
- `burnWithPermit(wrappedAsset: address, receiver: bytes, amount: uint256, deadline: uint256, v, r ,s)`

Argument | Description
---------- | ----------
**wrappedAsset** | The corresponding wrapped asset to burn. Must be the same as the `tokenContract` used in the `createPermit` function
**receiver** | The Hedera account to receive the wrapped tokens
**amount** | The amount of wrapped tokens to be burned and transferred
**deadline**: | Timestamp of the deadline
**v, r, s** | Information about the signature, computed when the user signs the permit.

_burnWithPermit_ works exactly as _burn_ but doesn't require submitting _approve_ TX before burning the tokens, but it is necessary that signature and deadline are provided. The user can use this function to do both operations in one step.

```typescript
const message: Permit = {
       owner: state.metamask!.selectedAddress(),
       spender: contractData.controller,
       value: amount,
       nonce: contractData.nonce,
       deadline // get latest block and add 1h in seconds
};

const domain: Domain = {
       name: contractData.name,
       version: "1",
       chainId: id!,
       verifyingContract: contractAddress
};

const data = JSON.stringify({
       types: {
           EIP712Domain,
           Permit
       },
       domain,
       primaryType: "Permit",
       message
});

const signature = signTypedV4Data(data);
burnWithPermit(tokenAddress, receiver!._toProto().serializeBinary(), amount, deadline, signature.v,signature.r,signature.s);
```

### Transaction verification

After the burn operation is completed a _burn_ event is fired which is captured by the validators. The event contains information about the burned amount and the receiver. After the validators capture the event they distribute the service fee and schedule a transaction to transfer the remaining amount to the receiving Hedera account. The _Bridge Account_ balance is used for the transfer of the burned tokens. The _Bridge Fee Account_ is used as a Hedera transaction fee payer account for the final transfer.
>Note: In the case when the collected fee can not be divided equally between the validators the remainder from the devision is transferred to the receiving Hedera account.

### Monitoring the transfer

The user can query the Validator API in order to get information on the Bridge transfer using the EVM `TX Hash` and the `logIndex` of the `burn` event that is emitted as part of the `burn` / `burnWithPermit` transactions.

    GET {validator_host}:{validator_port}/api/v1/events/{burn_event_id}

where `burn_event_id` is the id of the Ethereum burn event. It must be constructed in the form:

`{TX-Hash}-{LogIndex}`

Parameter| Description
------ | -------
txHash | Transaction hash of the `burn` or `burnWithPermit` transactions
logIndex | Index of the burn event in the transaction receipt

Example format: TODO

TODO add info on what is the response from the Validator API