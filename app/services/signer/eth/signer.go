package eth

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	log "github.com/sirupsen/logrus"
)

type Signer struct {
	privateKey *ecdsa.PrivateKey
}

func NewEthSigner(privateKey string) *Signer {
	pk, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		log.Fatal(fmt.Sprintf("Invalid Ethereum Private Key provided: [%s]", privateKey))
	}
	return &Signer{privateKey: pk}
}

func (s *Signer) Sign(msg []byte) ([]byte, error) {
	return crypto.Sign(msg, s.privateKey)
}
