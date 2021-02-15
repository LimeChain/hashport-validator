package crypto_transfer

import (
	"testing"

	ethclient "github.com/limechain/hedera-eth-bridge-validator/app/clients/ethereum"
	hederaClients "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/signer/eth"
	"github.com/limechain/hedera-eth-bridge-validator/app/test/mocks/repository"
	"github.com/limechain/hedera-eth-bridge-validator/config"
)

func Test_CryptoTransferHandler(t *testing.T) {
	var cthConfig config.CryptoTransferHandler
	cthConfig.TopicId = "0.0.125563"
	cthConfig.PollingInterval = 5
	var ethClientConfig config.Ethereum
	ethClientConfig.NodeUrl = "wss://ropsten.infura.io/ws/v3/8b64d65996d24dc0aae2e0c6029e5a9b"
	ethClientConfig.BridgeContractAddress = "0x5eD8BBE6462B32B7Be746561918CDA3b9E985B2a"
	var ethPk string = "bb9282ba72b55a531fa5e7cc83e92e9055c6905648d673f4d57ad663a317da49"
	var apiAddress string = "http://testnet.mirrornode.hedera.com/api/v1/"
	var hederaCl config.Client
	hederaCl.NetworkType = "testnet"
	hederaCl.Operator.AccountId = "0.0.99661"
	hederaCl.Operator.PrivateKey = "302e020100300506032b657004220420fc79b49c62c4637437292814eccd640a6173933c30698ddb41c36c195f0b6629"
	hederaCl.Operator.EthPrivateKey = "bb9282ba72b55a531fa5e7cc83e92e9055c6905648d673f4d57ad663a317da49"
	var transactionRepo repository.MockTransactionRepository

	ethSigner := eth.NewEthSigner(ethPk)
	ethClient := ethclient.NewEthereumClient(ethClientConfig)
	hederaMirrorClient := hederaClients.NewHederaMirrorClient(apiAddress)
	hederaNodeClient := hederaClients.NewNodeClient(hederaCl)

	ctHandler := NewCryptoTransferHandler(cthConfig, ethSigner, hederaMirrorClient, hederaNodeClient, transactionRepo, ethClient)
	payload := []byte{10, 30, 48, 46, 48, 46, 57, 57, 54, 54, 49, 45, 49, 54, 49, 51, 51, 56, 57, 49, 50, 49, 45, 49, 51, 53, 54, 51, 49, 52, 54, 49, 18, 42, 48, 120, 55, 99, 70, 97, 101, 50, 100, 101, 70, 49, 53, 100, 70, 56, 54, 67, 102, 100, 65, 57, 102, 50, 100, 50, 53, 65, 51, 54, 49, 102, 49, 49, 50, 51, 70, 52, 50, 101, 68, 68, 24, 144, 78, 34, 13, 49, 49, 50, 54, 50, 50, 49, 50, 51, 55, 50, 49, 49}

	ctHandler.Handle(payload)

	//fmt.Println(ctHandler)
}
