package eth

import (
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/crypto"
	log "github.com/sirupsen/logrus"
)

type Signer struct {
	privateKey *ecdsa.PrivateKey
}

func NewEthSigner(privateKey string) *Signer {
	pk, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		log.Fatalf("Invalid Ethereum Private Key provided: [%s]", privateKey)
	}
	return &Signer{privateKey: pk}
}

func (s *Signer) Sign(msg []byte) ([]byte, error) {
	signature, err := crypto.Sign(msg, s.privateKey)
	if err != nil {
		return nil, err
	}
	// note: https://github.com/ethereum/go-ethereum/issues/1975
	signature[64] += 27

	return signature, nil
}
