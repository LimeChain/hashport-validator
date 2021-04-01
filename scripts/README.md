1. Run bridge-setup.go with privateKey, accountId and network as flags to generate the configurations

    `go run ./bridge-setup --privateKey=/your private key/ --accountId=/your account id/ --network=/previewnet|testnet|mainnet/ --members=/int, the count of the wanted bridge custodians/`

2. Run create-token.go to create custom token and associate it with hedera
   `go run ./token-create/ --/your private key/ --accountId=/your account id/ --network=/previewnet|testnet|mainnet/--memberPrKeys=/'The array of private keys from from the output of the previous step separated only by whitespace'/ --bridgeID=/The bridge id from the output of the previous step/`
