package coin_gecko

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
)

type SimplePriceResponse map[string]PriceResult

type PriceResult struct {
	Usd float64 `json:"usd"`
}

func ParseGetSimplePriceResponse(responseBody []byte) (result SimplePriceResponse, err error) {
	err = json.Unmarshal(responseBody, &result)
	if err != nil {
		log.Errorf("Error while parsing CoinGecko Simple Price response Body. Error: [%v]", err)
	}

	return result, err
}
