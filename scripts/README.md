1. Run bridge-setup.go with privateKey, accountId and network as flags to generate the configurations
    `go run ./scripts/bridge/setup.go --privateKey=/your private key/ --accountID=/your account id/ --adminKey=/your admin key/ --network=/previewnet|testnet|mainnet/ --members=/int, the count of the wanted bridge custodians/`

2. Run create.go to create custom token and associate it with hedera
   `go run ./scripts/token/native/create.go --privateKey=/your private key/ --accountID=/your account id/ --network=/previewnet|testnet|mainnet/ --memberPrKeys=/'The array of private keys from from the output of the previous step separated by ","'/ --bridgeID=/The bridge id from the output of the previous step/`

2. Run wrapped-token-create.go to create custom wrapped token with a bridge account treasury and associate it with hedera
   `go run ./scripts/token/wrapped/wrapped-token-create.go --privateKey=/your private key/ --accountID=/your account id/ --adminKey=/your admin key/ --network=/previewnet|testnet|mainnet/ --memberPrKeys=/'The array of private keys from from the output of the previous step separated by ","'/ --bridgeID=/The bridge id from the output of the previous step/ --generateSupplyKeysFromMemberPrKeys=true`

3. Associate new account to token
   `go run ./scripts/token/associate/associate.go --privateKey=/your private key/ --accountID=/your account id/ --network=/previewnet|testnet|mainnet/ --tokenID=/The Token id from the output of the previous step/`
