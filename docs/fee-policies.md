# Fee policies

Fee policies provide special fees to be applied for users from specific `Legal Entity`. Usually the fee policy will apply lesser fee for bridge transactions.
To enable fee policies to the bridge - submit _Topic Message_ with the _Fee Policy Config_ to `bridge.config_topic_id` from `config/bridge.yml` (See [Configuration](configuration.md), field `bridge.config_topic_id`).

## Terminology

During this document and in the code - the following terminology and meaning are used:

* `Legal Entity` - Legal entity for which the `Fee Policy` is applied. The actual legal entity is used only for logically separation. Subject to the `Fee Policy` are the defined list of `User Address` items.
* `User Address` - _EVM Wallet Address_ or _Hedera Account Id_ to which the `Fee Policy` will be applied.
* `Token Address` - _EVM Token Address_ or _Hedera Token Id_ used in bridge transaction for which the `Fee Policy` will be applied.
* `Target Network` - Network Id target of the bridge transaction, for which the `Fee Policy` will be applied. If it is not defined - no restrictions by network will be applied.
* `Fee Policy` - Fee policy definition, describing the actual fee calculation. The fee policy is an implementation of interface `FeePolicyInterface` in `app/model/fee-policy/fee_policy.go`.
* `Fee Policy Type` - The type describes the calculation logic of the fee. Possible values are `flat`, `percentage`, `flat_per_token`.

## Fee Policy Config structure

Available `Fee Policy Type` items are:

* Flat - code usage `constants.FeePolicyTypeFlat` - `Fee Policy` with specific flat fee without other restrictions.
* Percentage - code usage `constants.FeePolicyTypePercentage` - `Fee Policy` with specific percentage fee without other restrictions.
* Flat Per Token - code usage `constants.FeePolicyTypeFlatPerToken` - `Fee Policy` with specific flat fee per token.

```YAML
policies:
  "Some LTD": # Map of `Legal Entity` items. There should be one 
    addresses: # List of `User Address` part of the `Legal Entity`. At least one address should be added.
      - "0.0.101" # `User Address` item
      - "0.0.102" # `User Address` item
      - "0.0.103" # `User Address` item
    policy: # Policy section
      fee_type: "flat" # Describes `Fee Policy Type`
      networks: # Describes `Target Network`. The section is not mandatory. If not defined - the policy is applied for all networks
        - 8001 # `Target Network` Id
      value: 2000 # Specific `Fee Policy` item. Actual definition depends on the implementation. See next section.
```

## Supported Fee Policies

### Flat Fee Policy

Implemented in `app/model/fee-policy/flat_fee_policy.go` - struct `FlatFeePolicy`.

Flat `Fee Policy` is described with only the `flat` fee amount in YAML section `value`. The fee amount used for the bridge transaction is the `value` itself.

```YAML
policies:
  # ... Other `Legal Entity` map item definitions
  "Some LTD":
    addresses:
      - "0.0.101"
      - "0.0.102"
      - "0.0.103"
    policy:
      fee_type: "flat"
      networks:
        - 8001
      value: 2000 # Flat fee value of the `Fee Policy`
```

### Percentage Fee Policy

Implemented in `app/model/fee-policy/percentage_fee_policy.go` - struct `PercentageFeePolicy`.
Percentage `Fee Policy` is described with only the `percentage` in YAML section `value`. The fee amount used for the bridge transaction is calculated with the percentage `value` and bridge precision `constants.FeeMaxPercentage`.

```YAML
policies:
  # ... Other `Legal Entity` map item definitions
  "One More LTD":
    addresses:
      - "0.0.201"
      - "0.0.202"
      - "0.0.203"
    policy:
      fee_type: "percentage"
      networks:
        - 8001
      value: 2000 # Percentage fee value of the `Fee Policy`
```

### Fee Per Token Policy

Implemented in `app/model/fee-policy/flat_fee_per_token_policy.go` - struct `FlatFeePerTokenPolicy`.
Fee per token `Fee Policy` is described by map between `Token Address` and flat fee value. The fee amount of the bridge transaction is used if the token in the transaction is defined in the map.

```YAML
policies:
  # ... Other `Legal Entity` map item definitions
  "Some Other LTD":
    addresses:
      - "0.0.104"
      - "0.0.105"
    policy:
      fee_type: 'flat_per_token'
      networks:
        - 8001
        - 8002
      value: # Structure of the 
        - { token: "0.0.3001", value: 2000 } # token is `Token Address`; value is the flat fee value
        - { token: "0.0.3002", value: 1000 }
```

## Notes

* One `User Address` can use only one `Fee Policy`.
* _Fee Policy Config Parser_ will not throw error if a user is defined in many `Legal Entity` items. Only the "last" processed `User Address` will be used.
* _Fee Policy Config Parser_ does not require and will not validate `Fee Policy` with high or low fee amount value.