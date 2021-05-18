# Integration with Hedera <-> EVM-chain bridge
The Bridge provides users with functionality to transfer HBAR or HTS tokens from Hedera to EVM-based chain or Wrapped HBAR and Wrapped Tokens from EVM-based chain to Hedera. The Bridge is operated by registered validators that provide signatures for every requested transfer. The transfer is processed when the majority of validators verify the transfer (supermajority).

## 1. Hedera -> EVM-chain

This functionality allows users to transfer HBAR or any HTS token supported by the bridge and receive a wrapped version of the asset on the EVM chain.

### 1.1 How it works
The user needs to submit a transaction to Hedera account that is controlled by the bridge. **The transfer must contain MEMO with a valid receiving EVM address.** For every transfer a service fee is charged, which is configurable by the validators.

Once the transaction is successfully mined by HCS, you have successfully deposited your requested transfer to the bridge account. The validators are notified that crypto or token transfer has occurred and start providing their signatures. Each of them publishes his signature to a _Hedera Topic_ used specifically for that purpose. As soon as the last necessary validator publishes his signature and supermajority is reached, a scheduled transaction is created that transfers the service fee to each Validator.

### 1.2 Checking status of the transfer
Users can check the status of the transfer by making a API call to any of the validators like this:

    GET VALIDATOR_HOST:PORT/transfers/TRANSACTION_ID HTTP/1.1

>**VALIDATOR_HOST:PORT** - Host and port of any chosen validator
**TRANSACTION_ID** - Transaction id of the transfer (Ex: ...)

The response is in JSON format and contains the following data:
```go
    Recipient 		string 		`json:"recipient"`
    RouterAddress 	string 		`json:"routerAddress"`
    Amount 			string 		`json:"amount"`
    NativeAsset 	string 		`json:"nativeAsset"`
    WrappedAsset 	string 		`json:"wrappedAsset"`
    Signatures 		[]string 	`json:"signatures"`
    Majority 		bool 		`json:"majority"`
```
Property | Description
---------- | ----------
**Recipient** | EVM address of the receiver
**RouterAddress** | Address of the router contract
**Amount** | Transfer amount
**NativeAsset** | Alias for the transferred asset
**WrappedAsset** | Alias for the wrapped asset
**Signatures** | Array of all provided signatures by the validators at the moment
**Majority** | True if supermajority is reached and the transfer can be completed

### 1.3 Claiming wrapped token
When supermajority is reached only one step remains: for the user to claim their Wrapped version of HBAR or HTS token. In order to do that, the user must sign and submit a **mint transaction** to the Bridge Router Contract.

The mint operation can be constructed using the following arguments:

	mint(transactionId: bytes , wrappedAsset: address, receiver: address, amount: UInt256, signatures: bytes[] )

Argument | Description
---------- | ----------
**transactionId** | The Hedera Transaction ID
**wrappedAsset** | The corresponding wrappedToken contract address
**receiver** | The address receiving the tokens
**amount** | The desired minting amount
**signatures** | The array of signatures from the members, authorising the operation

### 1.4 Service fee distribution

The main incentive for the Validators is the `service fee` paid by users. The fee is a percentage of the transferred amount, paid on the native asset. The Service fee is configurable property and determined by the validators.

Fees are payed out from the Bridge account. Each Validator creates a Scheduled transaction and transfers the `service fee` amount from the Bridge account to the list of validators equally. Due to the nature of Scheduled Transactions, only one will be successfully executed, creating a scheduled Entity and all others will fail with `IDENTICAL_SCHEDULE_ALREADY_CREATED` error, and the transaction receipt will include the `ScheduleID` of the first submitted transaction. All validators, except the one that successfully created the Transaction execute `ScheduleSign` and once `n out of m` validators execute the Sign operation, the transfer of the fees will be executed.


## 2. EVM-chain -> Hedera
This functionality allows the user to transfer Wrapped HBAR or any supported by the bridge Wrapped Tokens from EVM-based chain to Hedera.

### 2.1 Query supported tokens

In order to get all the supported wrapped tokens by the bridge the user must do two things:

1. Get the wrapped tokens count from the Router contract by calling the function `wrappedAssetsCount()`.
2. Call the Router contract function `wrappedAssetAt(uint256  index)` for every value between 0 ... wrappedAssetsCount-1. Each time the function will return the **address** of the ERC20 wrapped token contract.

In order to get the corresponding Hedera native asset, one can query the following mapping:

`mapping(address => bytes) public wrappedToNative;`

where  `address`  represents the ERC20 address of the wrapped asset and  `bytes`  represent the HTS Entity ID or simply  `HBAR`  (in the case for HBAR-s).

### 2.1 Burn the wrapped asset
Transfer from EVM-chain to Hedera is achieved by submitting a **burn operation** to the Router Contract.
There are two supported contract functions by which this can be done:

- `burn(amount: uint256, receiver: bytes, wrappedAsset: address)`

Argument | Description
---------- | ----------
**amount** | The amount of wrapped tokens to be bridged
**receiver** | The Hedera account to receive the wrapped tokens
**wrappedAsset** | The corresponding wrapped asset contract address

In order to call the _burn_ function first the user must **permit** the operation to be executed by the Router Contract. This can be done by calling the _permit_ function in the corresponding _wrappedAsset_.
```
function permit(address owner, address spender, uint256 amount, uint256 deadline, uint8 v,bytes32 r, bytes32 s)
```
Argument | Description
---------- | ----------
**owner** | Address of token owner
**spender** | Router contract address
**amount** | The amount of wrapped tokens to be bridged
**deadline**: | Timestamp of the deadline
**v, r, s** | Information about the signature

Here is a example of how to create the necessary signature for permit operation:

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


- `burnWithPermit(wrappedAsset: address, receiver: bytes, amount: uint256, deadline: uint256, v, r ,s)`

Argument | Description
---------- | ----------
**wrappedAsset** | The corresponding wrapped asset contract address
**receiver** | The Hedera account to receive the wrapped tokens
**amount** | The amount of wrapped tokens to be bridged
**deadline**: | Timestamp of the deadline
**v, r, s** | Information about the signature

_burnWithPermit_ works exactly as _burn_ but doesn't require to submit _permit_ operation before burning the tokens, but it is necessary that signature and deadline are provided. The user can use this function to do both operations in one step.

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

After the burn operation is completed a _burn_ event is fired which is captured by the validators. The event contains information about the burned amount and the receiver. After the validators capture the event they distribute the service fee and schedule a transaction to transfer the remaining amount to the receiving Hedera account.
>Note: In the case when the collected fee can not be divided equally between the validators the remainder from the devision is transferred to the receiving Hedera account.

### 2.2 Monitoring the Transfer

The user can receive information about any scheduled transaction from Hedera mirror node. In order to do that the user needs the `SCHEDULED_TRANSACTION_ID`. The ID if the transaction can be retrieved by querying any of the validators at

    GET VALIDATOR_HOST:PORT/api/v1/events/BURN_EVENT_ID HTTP/1.1

where `BURN_EVENT_ID` is the id of the Ethereum burn event. It must be constructed in the form: `txHash-logIndex`

Parameter| Description
------ | -------
txHash | Transaction hash
logIndex | Index of the burn event in the transaction receipt

>Note: In order to return the scheduled transaction id, the event of course would need to be processed by the validators

Having the _scheduled transaction id_ the user can query any Hedera mirror node to get information about the transfer.
