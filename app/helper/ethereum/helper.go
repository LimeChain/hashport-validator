package ethereum

import "github.com/ethereum/go-ethereum/crypto"

var (
	LogEventItemSetName    = "ItemSet"
	logEventItemSetSig     = []byte("ItemSet(uint256,address)")
	LogEventItemSetSigHash = crypto.Keccak256Hash(logEventItemSetSig)
)
