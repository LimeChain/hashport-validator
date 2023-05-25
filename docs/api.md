# API
List of supported endpoints by the application:



- `GET /api/v1/config/bridge`: Returns as JSON object the full configuration of the [bridge.yml](configuration.md) where the keys are in `camelCase` format.
- `GET /api/v1/min-amounts`: Returns as JSON object the current min-amounts per asset per network in the following format:
```json
{
  "networkId": {
    "assetIdOrAddress": "min-amount"
  }
}
```
Example:
```json
{
  "295": {
    "HBAR": "20736132711",
    "0.0.26056684": "144956212352"
  }, 
  "1": {
    "0x14ab470682Bc045336B1df6262d538cB6c35eA2A": "20736132711",
    "0xac3211a5025414Af2866FF09c23FC18bc97e79b1": "1449562123521537231600"
  }, 
  "137": {
    "0x1646C835d70F76D9030DF6BaAeec8f65c250353d": "20736132711"
  }
}
```
- `POST /api/v1/transfers/history`: Accepts a request body in the form (`*` is required) and returns:
  - Maximum page size is 50. Pages start from 1.
  - Parameter timestamp supports query params like `gt`, `lt`, `gte`, `lte`, `eq` to filter by range.
  - ```json
    {
      *"page": 1,
      *"pageSize": 20,
      "filter": {
        "originator": "Hedera account ID or EVM address",
        "timestamp": "VALID RFC3339(Nano) DATE. Supports query params. Ex: 2021-08-31T00:00:00.000000000Z. Ex-2: gte=2023-05-25T07:43:08.650830003Z&lte=2023-05-25T08:11:10.058833356Z",
        "tokenId": "Hedera Token ID or EVM address",
        "transactionId": "Hedera Transaction ID or EVM transaction hash"
      }
    }
    ```
  - ```json
    {
      "items": [],
      "totalCount": 0
    }
    ```

- `GET /fees/nft`: Returns the fees for porting/burning NFT assets grouped by network. Ex:
- ```json
  {
    "295": {
      "tokenId or address": {
        "isNative": true,
        "paymentToken": "HBAR or address of the payment token",
        "fee": "fee amount"
      }
    },
  ...
  }
  ```

- `POST /transfer-reset`: Updates the stuck transfers to `COMPLETE` and `user_get_his_token` to 1
- ```bash
  curl --location --request POST 'http://localhost:9200/api/v1/transfer-reset' \
  --header 'Content-Type: application/json' \
  --data-raw '{
      "transactionId": "0.0.3121456-1680613460-129693178",
      "sourceChainId": 296,
      "targetChainId": 80001,
      "sourceToken": "HBAR",
      "Password": "passwordTestValidator"
  }'
  ```