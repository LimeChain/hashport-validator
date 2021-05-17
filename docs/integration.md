# Integration with Hedera <-> EVM-chain bridge
The Bridge provides users with functionality to transfer HBAR or HTS tokens from Hedera to EVM-based chain or Wrapped HBAR and Wrapped Tokens from EVM-based chain to Hedera. The Bridge is operated by registered validators that provide signatures for every requested transfer. The transfer is processed when the majority of validators verify the transfer (supermajority).

## 1. Hedera -> EVM-chain

This functionality allows users to transfer HBAR or any HTS token supported by the bridge and receive a wrapped version of the asset on the EVM chain. 

### 1.1 How it works
The user needs to submit a transaction to Hedera account that is controlled by the bridge. **The transfer must contain MEMO with a valid receiving EVM address.** For every transfer a service fee is charged, which is configurable by the validators.

Once the transaction is successfully mined by HCS, you have successfully deposited your requested transfer to the bridge account. The validators are notified that crypto or token transfer has occurred and start providing their signatures. Each of them publishes his signature to a _Hedera Topic_ used specifically for that purpose. As soon as the last necessary validator publishes his signature and supermajority is reached, a scheduled transaction is created that transfers the service fee to the Bridge Fee Account.

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

### 2.1 Supported wrapped tokens

In order to get all the supported wrapped tokens by the bridge the user must do two things:

 1. Get the wrapped tokens count from the Router contract by calling the function `wrappedAssetsCount()`.
 2. Call the Router contract function `wrappedAssetAt(uint256  index)` for every value between 0 ... wrappedAssetsCount-1. Each time the function will return the address of the ERC20 wrapped token contract.

If the user needs the name of the wrapped token he can retrieve it from public mapping in the Router contract by the ERC20 address:

	mapping(address => bytes) public wrappedToNative;

### 2.1 How it works
Transfer from EVM-chain to Hedera is achieved by submitting a **burn operation** to the Router Contract. 
There are two supported contract functions by which this can be done:

 - `burn(amount: uint256, receiver: bytes, wrappedAsset: address)`

Argument | Description
---------- | ----------
**amount** | The amount of wrapped tokens to be bridged
**receiver** | The Hedera account to receive the wrapped tokens
**wrappedAsset** | The corresponding wrapped asset contract address

After the burn operation is completed a _burn_ event is fired which is captured by the validators. The event contains information about the burned amount and the receiver. After the validators capture the event they distribute the service fee and schedule a transaction to transfer the remaining amount to the receiving Hedera account.


 - `burnWithPermit(wrappedAsset: address, receiver: bytes, amount: uint256, deadline: uint256, v, r ,s)`

Argument | Description
---------- | ----------
**wrappedAsset** | The corresponding wrapped asset contract address
**receiver** | The Hedera account to receive the wrapped tokens
**amount** | The amount of wrapped tokens to be bridged
**deadline**: | Timestamp of the deadline
**v, r, s** | Information about the signature

### 2.2 Getting information about the scheduled transaction



```go    
     message ScheduleGetInfoQuery {  
	     QueryHeader header = 1; // Standard query metadata, including payment
	     ScheduleID schedule = 2; // The id of an existing schedule
     }  
```
Response from query is in the following form:

```go
     message ScheduleGetInfoResponse {  
	     ScheduleID scheduleID = 1; // The id of the schedule
	     Timestamp deletionTime = 2; // If the schedule has been deleted, the consensus time when this occurred
	     Timestamp executionTime = 3; // If the schedule has been executed, the consensus time when this occurred
	     Timestamp expirationTime = 4; // The time at which the schedule will expire
     }
          
     SchedulableTransactionBody scheduledTransactionBody = 5; // The scheduled transaction
	     string memo = 6; // The publicly visible memo of the schedule
	     Key adminKey = 7; // The key used to delete the schedule from state
	     KeyList signers = 8; // The Ed25519 keys the network deems to have signed the scheduled transaction
	     AccountID creatorAccountID = 9; // The id of the account that created the schedule
	     AccountID payerAccountID = 10; // The id of the account responsible for the service fee of the scheduled transaction
	     TransactionID scheduledTransactionID = 11; // The transaction id that will be used in the record of the scheduled transaction (if it executes)
    }`
```
