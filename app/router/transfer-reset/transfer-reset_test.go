package transfer_reset

import (
	"testing"

	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"github.com/stretchr/testify/assert"
)

var (
	transferIdUrlParamKey = "id"
	transferId            = "1"
	transfer              = service.TransferData{}
)

func Test_NewRouter(t *testing.T) {
	router := NewRouter(mocks.MTransferService, mocks.MPrometheusService)

	assert.NotNil(t, router)
}
