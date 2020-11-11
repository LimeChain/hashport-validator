package hedera

import (
	hederasdk "github.com/hashgraph/hedera-sdk-go"
	config "github.com/limechain/hedera-eth-bridge-validator/config"
)

func NewMirrorClient(mirrorConfig config.MirrorNode) (hederasdk.MirrorClient, error) {
	return hederasdk.NewMirrorClient(mirrorConfig.Client)
}
