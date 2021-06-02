### Hedera to Ethereum transfers require the following transactions per validator:

- 1x  `ScheduleCreate` ->Transfer of fees: For `n` validators, `n-1` will fail and one will be successful. The cost of
  this transaction depends slightly on the validators count:

n | Estimated transaction cost
------ | ------  
3 Validators | 0.0102$ or 0.04365ℏ
7 Validators | 0.0104$ or 0.044585ℏ

Estimations are calculated by [Hedera Fee Calculator](https://hedera.com/fees) and are based on the following
parameters:

Parameter | Value
------ | -------  
payer sigs | 1
byte size | 304 bytes _(for 3 validators)_
scheduled txn sigs | 1
expiration | 1 hour
transaction memo size | 31 bytes
admin keys | 1
Entity memo size | 0 bytes
Non-scheduling payer? | Yes

- 1x  `ScheduleSign` -> transfer of fees. Submitted if the  `ScheduleCreate` transaction failed.

**Estimated transaction cost**: 0.001$ or 0.00425ℏ.

The estimation is based on the following parameters:

Parameter | Value
------ | -------  
payer sigs | 1
scheduled txn sigs | 1
expiration | 1 hour

- 1x  `ConsensusMessageSubmit` -> Submitting the authorisation signature

**Estimated transaction cost**: 0.0001$ or 0.000467ℏ.

The estimation is based on the following parameters:

Parameter | Value
------ | -------  
payer sigs | 1
total sigs | 1
Size of message | 309
Transaction memo size | 0 bytes

> Total fee that needs to be payed by each validator (in the case of 3 validators): **0.048367ℏ**

#### Mint Operation

In order for the user to receive wrapped tokens he must submit `mint` transaction to the contract.

> Estimated gas needed for this operation: **107500 GWei**
Estimated tx fee needed for this transaction: **0.00107 MATIC**

### Ethereum to Hedera transfers require the following transactions per validator:

- 1x  `ScheduleCreate` ->Transfer of fees: Like the case of transfer to Ethereum for `n` validators, `n-1` will fail and
  one will be successful. The cost of this transaction depends slightly on the validators count:

n | Estimated transaction cost
------ | ------  
3 Validators | 0.0104$ or 0.044585ℏ
7 Validators | 0.0106$ or 0.045222ℏ

Estimations are calculated by Hedera Fee Calculator and are based on the following parameters:

Parameter | Value
------ | -------  
payer sigs | 1
byte size | 410 bytes _(for 3 Validators)_
scheduled txn sigs | 1
expiration | 1 hour
transaction memo size | 68 bytes
admin keys | 1
Entity memo size | 0 bytes
Non-scheduling payer? | Yes

- 1x  `ScheduleSign` -> Transfer of fees:  Behaves exactly like transfer to Ethereum and the fee estimation is the same.

> Total fee that needs to be payed by each validator (in the case of 3 validators): **0.048835ℏ**

#### Burn Operation

In order for the user to receive HTS tokens he must submit `burnWithPermit` transaction to the contract.

> Estimated gas needed for this operation: **76900 GWei**  
Estimated tx fee needed for this transaction:  **0.000768 MATIC**
