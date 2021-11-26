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
>Memo: {targetChainId}-0x700d8a76b37f672a06ab89fe1ec95acfba799f1c

The corresponding `TransactionID` is a unique identifier for the Bridge transfer operation and can be used to query the status of the Bridge transfer.

### Step 2. Waiting for Signatures

Bridge operators (validators) submit their signatures to an `HCS` topic, however for the ease of use, clients can query the signatures for a given transfer directly from the Validator's API:

    GET {validator_url}:{port}/api/v1/transfers/{transaction_id}

Where `transaction_id` is the Hedera `TransactionID` of the `CryptoTransfer` sending the asset to the Bridge account.

The response is in JSON format and contains the following data:

```json
{
  "isNft": false,
  "recipient": "0x700d8a76b37f672a06ab89fe1ec95acfba799f1c",
  "routerAddress": "0x",
  "amount": "100",
  "sourceChainId": "",
  "targetChainId": "",
  "sourceAsset": "",
  "nativeAsset": "",
  "targetAsset": "",
  "signatures": [
  ],
  "majority": false
}
```

Property | Description
---------- | ----------
**isNft** | Whether the transfer is fungible or non-fungible
**recipient** | EVM address of the receiver
**routerAddress** | Address of the router contract
**amount** | The transfer amount minus the services fee that is applied. If service fee is 1% and original transfer amount is 100 Hbars, the returned property will have 99 hbars.
**sourceChainId** | The chain ID from which the transfer has been initiated
**targetChainId** | The chain ID to which the transfer data and has to be submitted
**sourceAsset** | The asset from the source chain
**nativeAsset** | The native asset of the transferred asset
**targetAsset** | The target asset for the transfer
**signatures** | Array of all provided signatures by the validators up until this moment
**majority** | True if supermajority is reached and the wrapped token may be claimed

### Step 3. Claiming Wrapped Asset

Once supermajority is reached the users can claim _wrapped version_ of the asset. In order to do that, the user must sign and submit a **mint transaction** to the Bridge Router Contract.

The mint operation can be constructed using the following arguments:

	mint(uint256 sourceChainId, bytes transactionId, address targetAsset, address recipient, uint256 amount, bytes[] signatures)

Argument | Description
---------- | ----------
**transactionId** | The Hedera `TransactionID` of the Deposit transaction. Converting the TX ID string to bytes:`Web3.utils.fromAscii(transactionId)`

### Service Fee

The main incentive for the Validators is the `service fee` charged on every transfer. The fee is a percentage of the transferred amount, paid on the native asset. The Service fee is configurable property and determined by the validators.
Fees are paid out from the Bridge account.


## Return Wrapped EVM assets to Native Hedera
This functionality allows the user to transfer Wrapped HBAR or any supported by the bridge Wrapped Tokens from EVM-based chain to Hedera.

### Step 1. Burn the Wrapped asset

Transfers from EVM chain to Hedera is achieved by submitting a **burn operation** to the Router Contract.
There are two supported contract functions by which this can be done:

#### Option 1 - Approve + Burn

The straightforward way for burning the wrapped assets would be by executing 2 transactions:
1. ERC20 Approve
2. Router Burn 

The burn transaction has the following format:
- `burn(uint256 targetChainId, address wrappedAsset, uint256 amount, bytes receiver)`

Argument | Description
---------- | ----------
**targetChainId** | The chain id to which you would like to bridge transfer. In the case of Hedera, it must be `0`
**wrappedAsset** | The corresponding wrapped asset to burn
**amount** | The amount of wrapped tokens to be burned and transferred in their native version on Hedera
**receiver** | The Hedera Account to receive the native representation of the wrapped asset

>Note: The receiver [AccountId](https://hashgraph.github.io/hedera-sdk-java/index.html?com/hedera/hashgraph/sdk/account/AccountId.html) must be serialized by Hedera SDK as such:
>`accountId._toProto().serializeBinary()`, before passing it as an argument to the _burn_ function.

#### Option 2 - Sign Permit + Burn

Using the Permit design, we are able to initiate the transfer of the assets in one transaction. Instead of the user executing a separate ERC20 `approve` TX, he must sign a `permit` message that is verified by the ERC20 contract in order to authorise the Router contract to spend the user's funds on his behalf.

Here is an example on how to create the necessary signature for permit operation:

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
            chainId: "", // chain id of the EVM network
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
- `burnWithPermit(uint256 targetChainId, address wrappedAsset, uint256 amount, bytes receiver, uint256 deadline, uint8 v, uint8 r ,uint8 s)`

Argument | Description
---------- | ----------
**targetChainId** | The chain id to which you would like to bridge transfer. In the case of Hedera, it must be `0`
**wrappedAsset** | The corresponding wrapped asset to burn. Must be the same as the `tokenContract` used in the `createPermit` function
**amount** | The amount of wrapped tokens to be burned and transferred
**receiver** | The Hedera account to receive the wrapped tokens
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
       chainId: "1", // Ethereum Mainnet ChainID
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
>Note: In the case when the collected fee can not be divided equally between the validators the remainder from the division is transferred to the receiving Hedera account.

### Monitoring the transfer

The user can query the Validator API in order to get information on the Bridge transfer using the EVM `TX Hash` and the `logIndex` of the `burn` event that is emitted as part of the `burn` / `burnWithPermit` transactions.

    GET {validator_host}:{validator_port}/api/v1/events/{burn_event_id}/tx

where `burn_event_id` is the id of the Ethereum burn event. It must be constructed in the form:

`{TX-Hash}-{LogIndex}`

Parameter| Description
------ | -------
txHash | Transaction hash of the `burn` or `burnWithPermit` transactions
logIndex | Index of the burn event in the transaction receipt

Example format: `0x00cf6cbfbfd1f48dbcdef5cf2ce982085422434ce9a8fd21246cb2f39de8a94a-14`
If the transfer is not processed yet, the response will be `404`. 
if the transfer has been processed, and the funds have been transferred, the `ScheduledTransaction ID` is returned. Using the Scheduled Transaction ID, users can query the Mirror node and see the details of the transfer  

## Wrap Native EVM Native assets on Hedera
This functionality allows the user to transfer native EVM Tokens to Hedera.

### Step 1. Lock the Native Asset
EVM Native to Hedera is achieved by submitting a **lock operation** to the Router Contract.
There are two supported contract functions by which this can be done:

### Option 1 - Approve + Lock

The straightforward way for locking the native assets would be by executing the following transactions:
1. ERC20 Approve
2. Router Lock

The lock transaction has the following format:
- `lock(uint256 targetChainId, address nativeAsset, uint256 amount, bytes receiver)`

Argument | Description
---------- | ----------
**targetChainId** | The chain id to which you would like to bridge transfer. In the case of Hedera, it must be `0`
**nativeAsset** | The corresponding native token to lock
**amount** | The amount of tokens to be locked and transferred in their wrapped version on Hedera
**receiver** | The Hedera Account to receive the wrapped representation of the native asset

>Note: The receiver [AccountId](https://hashgraph.github.io/hedera-sdk-java/index.html?com/hedera/hashgraph/sdk/account/AccountId.html) must be serialized by Hedera SDK as such:
>`accountId._toProto().serializeBinary()`, before passing it as an argument to the _lock_ function.

#### Option 2 - Sign Permit + Lock

`NB!` Native tokens must support permit in order this option to be viable.

Using the Permit design, we are able to initiate the transfer of the assets in one transaction. 
Instead of the user executing a separate ERC20 `approve` TX, he must sign a `permit` message that is verified by the 
ERC20 contract in order to authorise the Router contract to spend the user's funds on his behalf.

> Note: **deadline** is the timestamp, after which the _permit_ for _lock operation_ will not be active. The user can set it however he likes.

> More information about the _permit_ operation can be found here: [EIP 2612](https://eips.ethereum.org/EIPS/eip-2612)

Once the user signs the `permit`, the `lockWithPermit` transaction can be executed. The signature (`v`, `r` and `s`) along with the `deadline` are send as part of the lock transaction:
- `lockWithPermit(uint256 targetChainId, address nativeAsset, uint256 amount, bytes receiver, uint256 deadline, uint8 v, uint8 r ,uint8 s)`

### Transaction verification

After the lock operation is completed a _lock_ event is fired which is captured by the validators. 
The event contains information about the locked amount and the receiver. 
After the validators capture the event they subtract the service fee and schedule `mint` and `transfer` transactions with the remaining amount to the receiving Hedera account.

### Monitoring the transfer

The user can query the Validator API in order to get information on the Bridge transfer using the EVM `TX Hash` and the `logIndex` of the `lock` event that is 
emitted as part of the `lock` / `lockWithPermit` transactions.

    GET {validator_host}:{validator_port}/api/v1/events/{lock_event_id}/tx

where `lock_event_id` is the id of the Ethereum Lock event. It must be constructed in the form:

`{TX-Hash}-{LogIndex}`

Parameter| Description
------ | -------
txHash | Transaction hash of the `lock` or `lockWithPermit` transactions
logIndex | Index of the lock event in the transaction receipt

## Return Wrapped Hedera assets to EVM Native

This functionality allows users to transfer wrapped HTS tokens supported by the bridge back to their original EVM chain.

### Step 1. Deposit Transaction

In order to initiate transfer, the user needs to submit a deposit transaction to the Hedera Bridge Account. More information on the account can be found in the [overview](./overview.md) document.

The transfer **must** specify the receiving address in the `memo` field.

Example:

>Transfer amount: 100 WrappedDAI (0.0.472374)
>Memo: {targetChainId}-0x700d8a76b37f672a06ab89fe1ec95acfba799f1c

where `targetChainId` is the chainId of the native network of the token. (Ethereum Mainnet - 1, Polygon Mainnet - 137)
`0x700d8a76b37f672a06ab89fe1ec95acfba799f1c` is the recipient.

The corresponding `TransactionID` is a unique identifier for the Bridge transfer operation and can be used to query the status of the Bridge transfer.

### Step 2. Waiting for Signatures

Bridge operators (validators) submit their signatures to an `HCS` topic, however for the ease of use, clients can query the signatures for a given transfer directly from the Validator's API:

    GET {validator_url}:{port}/api/v1/transfers/{transaction_id}

Where `transaction_id` is the Hedera `TransactionID` of the `CryptoTransfer` sending the asset to the Bridge account.

The response is in JSON format and contains the following data:

```json
{
  "isNft": false,
  "recipient": "0x700d8a76b37f672a06ab89fe1ec95acfba799f1c",
  "routerAddress": "0x",
  "amount": "100",
  "sourceChainId": "",
  "targetChainId": "",
  "sourceAsset": "",
  "nativeAsset": "",
  "targetAsset": "",
  "signatures": [
  ],
  "majority": false
}
```

Property | Description
---------- | ----------
**isNft** | Whether the transfer is fungible or non-fungible
**recipient** | EVM address of the receiver
**routerAddress** | Address of the router contract
**amount** | The burned amount
**sourceChainId** | The chain ID from which the transfer has been initiated
**targetChainId** | The chain ID to which the transfer data and has to be submitted
**sourceAsset** | The asset from the source chain
**nativeAsset** | The native asset of the transferred asset
**targetAsset** | The target asset for the transfer
**signatures** | Array of all provided signatures by the validators up until this moment
**majority** | True if supermajority is reached and the wrapped token may be claimed

### Step 3. Unlock the Native Asset

Once supermajority is reached users can unlock the _native version_ of the asset. In order to do that, the user must submit an **unlock transaction** to the Bridge Router Contract.

The `unlock` operation can be constructed using the following arguments:

	unlock(uint256 sourceChainId, bytes transactionId, address nativeAsset, uint256 amount, address recipient, bytes[] signatures)

Argument | Description
---------- | ----------
**transactionId** | The Hedera `TransactionID` of the Deposit transaction. Converting the TX ID string to bytes:`Web3.utils.fromAscii(transactionId)`


## Wrap EVM Native assets to another EVM

### Step 1. Lock the Native Asset
EVM Native to Hedera is achieved by submitting a **lock operation** to the Router Contract.
There are two supported contract functions by which this can be done:

### Option 1 - Approve + Lock

The straightforward way for locking the native assets would be by executing the following transactions:
1. ERC20 Approve
2. Router Lock

The lock transaction has the following format:
- `lock(uint256 targetChainId, address nativeAsset, uint256 amount, bytes receiver)`

Argument | Description
---------- | ----------
**targetChainId** | The chain id to which you would like to bridge transfer
**nativeAsset** | The corresponding native token to lock
**amount** | The amount of tokens to be locked and transferred in their wrapped version on the other EVM
**receiver** | The address to receive the wrapped representation of the native asset

#### Option 2 - Sign Permit + Lock

`NB!` Native tokens must support permit in order this option to be viable.

Using the Permit design, we are able to initiate the transfer of the assets in one transaction.
Instead of the user executing a separate ERC20 `approve` TX, he must sign a `permit` message that is verified by the
ERC20 contract in order to authorise the Router contract to spend the user's funds on his behalf.

> Note: **deadline** is the timestamp, after which the _permit_ for _lock operation_ will not be active. The user can set it however he likes.

> More information about the _permit_ operation can be found here: [EIP 2612](https://eips.ethereum.org/EIPS/eip-2612)

Once the user signs the `permit`, the `lockWithPermit` transaction can be executed. The signature (`v`, `r` and `s`) along with the `deadline` are send as part of the lock transaction:
- `lockWithPermit(uint256 targetChainId, address nativeAsset, uint256 amount, bytes receiver, uint256 deadline, uint8 v, uint8 r ,uint8 s)`

### Transaction verification

After the lock operation is completed a _lock_ event is fired which is captured by the validators.
The event contains information about the locked amount and the receiver.

### Step 2. Waiting for Signatures

Bridge operators (validators) submit their signatures to an `HCS` topic, however for the ease of use, clients can query the signatures for a given transfer directly from the Validator's API:

    GET {validator_url}:{port}/api/v1/transfers/{lock_event_id}

where `lock_event_id` is the id of the Ethereum Lock event. It must be constructed in the form:

`{TX-Hash}-{LogIndex}`

Parameter| Description
------ | -------
txHash | Transaction hash of the `lock` or `lockWithPermit` transactions
logIndex | Index of the lock event in the transaction receipt

The user can query the Validator API in order to get information on the Bridge transfer using the EVM `TX Hash` and the `logIndex` of the `lock` event that is
emitted as part of the `lock` / `lockWithPermit` transactions.

The response is in JSON format and contains the following data:

```json
{
  "isNft": false,
  "recipient": "0x700d8a76b37f672a06ab89fe1ec95acfba799f1c",
  "routerAddress": "0x",
  "amount": "100",
  "sourceChainId": "",
  "targetChainId": "",
  "sourceAsset": "",
  "nativeAsset": "",
  "targetAsset": "",
  "signatures": [
  ],
  "majority": false
}
```

Property | Description
---------- | ----------
**isNft** | Whether the transfer is fungible or non-fungible
**recipient** | EVM address of the receiver
**routerAddress** | Address of the router contract
**amount** | The transfer original transfer amount minus the services fee that is applied. If service fee is 1% and original transfer amount is 100 Hbars, the returned property will have 99 hbars.
**sourceChainId** | The chain ID from which the transfer has been initiated
**targetChainId** | The chain ID to which the transfer data and has to be submitted
**sourceAsset** | The asset from the source chain
**nativeAsset** | The native asset of the transferred asset
**targetAsset** | The target asset for the transfer
**signatures** | Array of all provided signatures by the validators up until this moment
**majority** | True if supermajority is reached and the wrapped token may be claimed

### Step 3. Claiming Wrapped Asset

Once supermajority is reached the users can claim _wrapped version_ of the asset. In order to do that, the user must sign and submit a **mint transaction** to the Bridge Router Contract.

The mint operation can be constructed using the following arguments:

	mint(uint256 sourceChainId, bytes transactionId, address targetAsset, address recipient, uint256 amount, bytes[] signatures)

Argument | Description
---------- | ----------
**transactionId** | `{TX-Hash}-{LogIndex}`


## Return Wrapped EVM assets to EVM Native
This functionality allows the user to transfer Wrapped EVM Tokens from one EVM chain to their Native EVM Chain.

### Step 1. Burn the Wrapped asset

Transferring back is achieved by submitting a **burn operation** to the Router Contract on the Wrapped EVM Network.
There are two supported contract functions by which this can be done:

#### Option 1 - Approve + Burn

The straightforward way for burning the wrapped assets would be by executing 2 transactions:
1. ERC20 Approve
2. Router Burn

The burn transaction has the following format:
- `burn(uint256 targetChainId, address wrappedAsset, uint256 amount, bytes receiver)`

Argument | Description
---------- | ----------
**targetChainId** | The chain id to which you would like to bridge transfer
**wrappedAsset** | The corresponding wrapped asset to burn
**amount** | The amount of wrapped tokens to be burned and unlocked on the `targetChainId`
**receiver** | The EVM address which will receive the native tokens on the `targetChainId`

#### Option 2 - Sign Permit + Burn

Using the Permit design, we are able to initiate the transfer of the assets in one transaction. Instead of the user executing a separate ERC20 `approve` TX, he must sign a `permit` message that is verified by the ERC20 contract in order to authorise the Router contract to spend the user's funds on his behalf.

Here is an example on how to create the necessary signature for permit operation:

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
            chainId: "", // chain id of the EVM network
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
- `burnWithPermit(uint256 targetChainId, address wrappedAsset, uint256 amount, bytes receiver, uint256 deadline, uint8 v, uint8 r ,uint8 s)`

Argument | Description
---------- | ----------
**targetChainId** | The chain id to which you would like to bridge transfer.
**wrappedAsset** | The corresponding wrapped asset to burn. Must be the same as the `tokenContract` used in the `createPermit` function
**amount** | The amount of wrapped tokens to be burned and transferred
**receiver** | The EVM address which will receive the native tokens on the `targetChainId`
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
       chainId: "1", // Ethereum Mainnet ChainID
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

After the burn operation is completed a _burn_ event is fired which is captured by the validators. 

### Monitoring the transfer

The user can query the Validator API in order to get information on the Bridge transfer using the EVM `TX Hash` and the `logIndex` of the `burn` event that is emitted as part of the `burn` / `burnWithPermit` transactions.

    GET {validator_url}:{port}/api/v1/transfers/{burn_event_id}

where `burn_event_id` is the id of the Ethereum Burn event. It must be constructed in the form:

`{TX-Hash}-{LogIndex}`

Parameter| Description
------ | -------
txHash | Transaction hash of the `burn` or `burnWithPermit` transactions
logIndex | Index of the burn event in the transaction receipt

```json
{
  "isNft": false,
  "recipient": "0x700d8a76b37f672a06ab89fe1ec95acfba799f1c",
  "routerAddress": "0x",
  "amount": "100",
  "sourceChainId": "",
  "targetChainId": "",
  "sourceAsset": "",
  "nativeAsset": "",
  "targetAsset": "",
  "signatures": [
  ],
  "majority": false
}
```

Property | Description
---------- | ----------
**isNft** | Whether the transfer is fungible or non-fungible
**recipient** | EVM address of the receiver
**routerAddress** | Address of the router contract
**amount** | The transfer original transfer amount minus the services fee that is applied
**sourceChainId** | The chain ID from which the transfer has been initiated
**targetChainId** | The chain ID to which the transfer data and has to be submitted
**sourceAsset** | The asset from the source chain
**nativeAsset** | The native asset of the transferred asset
**targetAsset** | The target asset for the transfer
**signatures** | Array of all provided signatures by the validators up until this moment
**majority** | True if supermajority is reached and the wrapped token may be claimed

### Step 3. Unlock the Native Asset

Once supermajority is reached users can unlock the _native version_ of the asset. In order to do that, the user must submit an **unlock transaction** to the Bridge Router Contract.

The `unlock` operation can be constructed using the following arguments:

	unlock(uint256 sourceChainId, bytes transactionId, address nativeAsset, uint256 amount, address recipient, bytes[] signatures)

Argument | Description
---------- | ----------
**transactionId** | `{TX-Hash}-{LogIndex}`


## NFT Transfers from Hedera to EVM
The steps below will showcase a bridge transfer from Hedera Native NFT to any EVM chain.

### Step 1. Deposit Transaction

In order to initiate transfer, user must submit a deposit transfer transaction, which will include the serial number of the given TokenID and an HBAR transfer with the fee.
Similar to fungible transfers, the transfer **must** include a memo, in the following format: `{targetChainId}-{receiverAddress}`, where:

* `targetChainId` - the chain id of the network to which the bridge transfer will go to.
* `receiverAddress` - the address to which the wrapped version of the NFT will be minted. 

Example of how to submit a Deposit transaction to the Bridge account in Go using Hedera Go SDK:
```go
nftID, _ := hedera.NftIDFromString("11@0.0.15633470") // 11 - Hedera Token serial number, 0.0.15633470 - tokenID of the Hedera token
response, _ := hedera.NewTransferTransaction().
    SetTransactionMemo("80001-0x0000000000000000000000000000000000000002"). // 80001 - chainId of Polygon Mumbai Testnet, 0x0000000000000000000000000000000000000002 - receiver address
    AddNftTransfer(nftID, client.GetOperatorAccountID(), bridgeAccount). // Send the NFT to the bridge account
    AddHbarTransfer(bridgeAccount, hedera.HbarFrom(1, "hbar")). // Send the NFT bridge transfer fee to the bridge account
    Execute(client)
```

The corresponding `response.TransactionID` must be converted in the following format: `{accountID}-{validStartSeconds}-{validStartNanos}`.
The transactionID is a unique identifier for the Bridge transfer operation and can be used to query the status of the Bridge transfer.

### Step 2. Waiting for Signatures

After the user has submitted the deposit transfer, he can begin to query a validator's API to see whether the necessary components for the transfer are met:

    GET {validator_url}/api/v1/transfers/{transaction_id}

Where `transaction_id` is the Hedera `TransactionID` of the NFT transfer to the Bridge account.

The response is in JSON format and contains the following data:

```json
{
  "isNft": true,
  "recipient": "0x700d8a76b37f672a06ab89fe1ec95acfba799f1c",
  "routerAddress": "0x",
  "sourceChainId": "",
  "targetChainId": "",
  "sourceAsset": "",
  "nativeAsset": "",
  "targetAsset": "",
  "tokenId": "",
  "metadata": "",
  "signatures": [
  ],
  "majority": false
}
```

Property | Description
---------- | ----------
**isNft** | Whether the transfer is fungible or non-fungible
**recipient** | EVM address of the receiver
**routerAddress** | Address of the router contract
**sourceChainId** | The chain ID from which the transfer has been initiated
**targetChainId** | The chain ID to which the transfer data and has to be submitted
**sourceAsset** | The asset from the source chain
**nativeAsset** | The native asset of the transferred asset
**targetAsset** | The target asset for the transfer
**tokenId** | The tokenId of the NFT
**metadata** | The metadata/tokenURI of the NFT
**signatures** | Array of all provided signatures by the validators
**majority** | True if supermajority is reached and the wrapped token may be claimed

### Step 3. Mint Wrapped Asset

Once majority is reached, users can mint the _wrapped version_ of the NFT. In order to do that, users must submit a `mintERC721` transaction to the Bridge Router Contract.

The mint operation can be constructed using the following arguments:

	mintERC721(uint256 sourceChainId, bytes transactionId, address targetAsset, uint256 tokenId, string metadata, address recipient, bytes[] signatures)

Argument | Description
---------- | ----------
**transactionId** | The Hedera `TransactionID` of the Deposit transaction. Converting the TX ID string to bytes for JS/TS: `Web3.utils.fromAscii(transactionId)` or `ethers.utils.toUtf8Bytes(transactionId)`.
**signatures** | Depending on the library chosen for EVM submission, it might be required to add to each signature a `0x` prefix.

## NFT Transfers from EVM back to Hedera
The steps below will showcase a bridge transfer of a wrapped NFT from any EVM back to Hedera.

### Step 1. Burn the Wrapped NFT Asset
Wrapped NFT burn consists of the following actions:
* An ERC-20 fee for the ERC-721 is sent from the user to the Router contract.
* An ERC-721 tokenId is burnt from the user.

Before user submits the burn transaction, two approvals will be needed:
* Approve ERC-20 bridge fee for Router contract:
  * To find the ERC-721's ERC-20 fee token, call the router contract with the following method:
  `const erc721PaymentTokenAddress = await routerContract.erc721Payment(address _wrappedERC721)`, [Router ABI](https://github.com/LimeChain/hashport-validator/blob/develop/app/clients/evm/contracts/router/diamond-router.go#L50).
  * To find the ERC-721's required ERC-20 fee, call the router contract with the following method:
  `const erc721Fee = await routerContract.erc721Fee(address _wrappedERC721)`, [Router ABI](https://github.com/LimeChain/hashport-validator/blob/develop/app/clients/evm/contracts/router/diamond-router.go#L50).
  * Approve the Router contract with the ERC-20 token fee amount:
  `await erc721PaymentTokenContract.approve(routerContractAddress, erc721Fee)` - A standard ERC-20 ABI can be used.
* Approve ERC-721 wrapped NFT for Router contract:
  `await erc721Contract.approve(routerContractAddress, tokenId)` - A standard ERC-721 ABI can be used.

Now that everything has been approved, users can execute a `burnERC721` transaction to the Router Contract.

	burnERC721(uint256 targetChainId, address erc721ContractAddress, uint256 tokenId, bytes receiver)

Argument | Description
---------- | ----------
**targetChainId** | The chain id of the Network to which the NFT is bridged. In the case of Hedera, it must be `0`.
**erc721ContractAddress** | The address of the wrapped ERC-721 Contract.
**tokenId** | The token ID to be bridged.
**receiver** | The Account/Address to receive the native representation of the wrapped asset. If the transfer is to Hedera, the receiver must be encoded in the SDK `hedera.AccountID.toBytes()` protobuf format.

### Step 2. Find the corresponding transaction
    GET {validator_url}/api/v1/transfers/{burn_erc721_id}

where `burn_erc721_id` is constructed in the following format: `{transactionHash}-{eventLogIndex}`

Argument | Description
--------- | ----------
**transactionHash** | The transaction hash of the `burnERC721` transaction
**eventLogIndex** | The log index of the `BurnERC721` event from the transaction receipt.