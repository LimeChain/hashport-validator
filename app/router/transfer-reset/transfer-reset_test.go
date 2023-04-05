package transfer_reset

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	transferModel "github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"github.com/stretchr/testify/assert"
)

var (
	transferId = "123"
)

func Test_NewRouter(t *testing.T) {
	router := NewRouter(mocks.MTransferService, mocks.MPrometheusService)

	assert.NotNil(t, router)
}

func Test_GetTransferSuccess(t *testing.T) {
	mocks.Setup()

	// Set up request body
	body := transferModel.TransferReset{
		TransactionId: transferId,
		SourceChainId: 1,
		TargetChainId: 2,
		SourceToken:   "token",
	}
	reqBody, _ := json.Marshal(body)

	mocks.MTransferService.On("UpdateTransferStatusCompleted", transferId).Return(nil)
	mocks.MPrometheusService.On("GetIsMonitoringEnabled").Return(false)

	handler := transferReset(mocks.MTransferService, mocks.MPrometheusService)

	req := httptest.NewRequest(http.MethodPost, "/transfer-reset", bytes.NewBuffer(reqBody))
	w := httptest.NewRecorder()
	handler(w, req)
	res := w.Result()
	defer res.Body.Close()
	data, _ := ioutil.ReadAll(res.Body)

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "OK", string(data))

}

func Test_GetTransferSuccess_FailsWithDB(t *testing.T) {
	mocks.Setup()

	// Set up request body
	body := transferModel.TransferReset{
		TransactionId: transferId,
		SourceChainId: 1,
		TargetChainId: 2,
		SourceToken:   "token",
	}
	reqBody, _ := json.Marshal(body)

	mocks.MTransferService.On("UpdateTransferStatusCompleted", transferId).Return(errors.New("error"))
	mocks.MPrometheusService.On("GetIsMonitoringEnabled").Return(false)

	handler := transferReset(mocks.MTransferService, mocks.MPrometheusService)

	req := httptest.NewRequest(http.MethodPost, "/transfer-reset", bytes.NewBuffer(reqBody))
	w := httptest.NewRecorder()
	handler(w, req)
	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func Test_GetTransferSuccess_FailsWithWrongParams(t *testing.T) {
	mocks.Setup()

	body := map[string]interface{}{
		"TransactionId": transferId,
		"SourceChainId": 1,
		"TargetChainId": "2",
		"SourceToken":   "token",
	}

	reqBody, _ := json.Marshal(body)

	mocks.MTransferService.On("UpdateTransferStatusCompleted", transferId).Return(nil)
	mocks.MPrometheusService.On("GetIsMonitoringEnabled").Return(false)

	handler := transferReset(mocks.MTransferService, mocks.MPrometheusService)

	req := httptest.NewRequest(http.MethodPost, "/transfer-reset", bytes.NewBuffer(reqBody))
	w := httptest.NewRecorder()
	handler(w, req)
	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}
