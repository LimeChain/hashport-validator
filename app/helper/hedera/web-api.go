package hedera

import (
	"encoding/base64"
	"encoding/json"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"strings"
)

func GetHederaFileRateFromResponseBody(responseBody []byte, log *log.Entry) (UpdatedFileRateData, error) {
	parsedResponse, err := parseHederaResponseFromResponseBody(responseBody, log)
	if err != nil {
		return UpdatedFileRateData{}, err
	}

	decodedMemo, err := base64.StdEncoding.DecodeString(parsedResponse.Transactions[0].MemoBase64)
	if err != nil {
		log.Errorf("Error while decoding MemoBase64 from HBAR Price Hedera Response. Error: [%v]", err)
		return UpdatedFileRateData{}, err
	}

	hederaFileRate := parseHederaFileRateFromMemo(decodedMemo, log)
	return hederaFileRate, nil
}

func parseHederaFileRateFromMemo(decodedMemo []byte, log *log.Entry) UpdatedFileRateData {
	result := UpdatedFileRateData{}
	asString := string(decodedMemo)
	allFieldsWithValues := strings.Split(asString, ", ")

	var err error
	for _, fieldWithValue := range allFieldsWithValues {
		fieldWithValueTokens := strings.Split(fieldWithValue, " : ")
		fieldName := fieldWithValueTokens[0]
		fieldValue := fieldWithValueTokens[1]

		switch fieldName {
		case "currentRate":
			result.CurrentRate, err = decimal.NewFromString(fieldValue)
			if err != nil {
				log.Errorf("Couldn't parse 'currentRate' from Hedera file for price rate. Error [%v]", err)
			}
		case "nextRate":
			result.NextRate, err = decimal.NewFromString(fieldValue)
			if err != nil {
				log.Errorf("Couldn't parse 'nextRate' from Hedera file for price rate. Error [%v]", err)
			}
		}
	}

	return result
}

func parseHederaResponseFromResponseBody(responseBody []byte, log *log.Entry) (result hederaResponse, err error) {
	err = json.Unmarshal(responseBody, &result)
	if err != nil {
		log.Errorf("Error while parsing HBAR response Body. Error: [%v]", err)
	}
	return result, err
}

type hederaResponse struct {
	Transactions []hederaTransaction `json:"transactions"`
	Links        map[string]string   `json:"links"`
}

type hederaTransaction struct {
	Bytes                    interface{}   `json:"bytes"`
	ChargedTxFee             int           `json:"charged_tx_fee"`
	ConsensusTimestamp       string        `json:"consensus_timestamp"`
	EntityId                 string        `json:"entity_id"`
	MaxFee                   string        `json:"max_fee"`
	MemoBase64               string        `json:"memo_base64"`
	Name                     string        `json:"name"`
	Node                     string        `json:"node"`
	Nonce                    int           `json:"nonce"`
	ParentConsensusTimestamp string        `json:"parent_consensus_timestamp"`
	Result                   string        `json:"result"`
	Scheduled                bool          `json:"scheduled"`
	TransactionHash          string        `json:"transaction_hash"`
	TransactionId            string        `json:"transaction_id"`
	Transfers                []interface{} `json:"transfers"`
	ValidDurationSeconds     string        `json:"valid_duration_seconds"`
	ValidStartTimestamp      string        `json:"valid_start_timestamp"`
}

type UpdatedFileRateData struct {
	CurrentRate decimal.Decimal
	NextRate    decimal.Decimal
}
