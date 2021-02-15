package exchangerate

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type ExchangeRateProvider struct {
	httpClient *http.Client
	rateURL    string
	coin       string
	currency   string
	rate       float64
}

func NewExchangeRateProvider(coin string, currency string) ExchangeRateProvider {
	return ExchangeRateProvider{
		httpClient: &http.Client{},
		rateURL:    fmt.Sprintf("https://api.coingecko.com/api/v3/simple/price?ids=%s&vs_currencies=%s", coin, currency),
		coin:       coin,
		currency:   currency,
	}
}

func (erp *ExchangeRateProvider) GetEthVsHbarRate() (float64, error) {
	response, err := erp.httpClient.Get(erp.rateURL)
	if err != nil {
		return 0, err
	}

	bodyBytes, err := readResponseBody(response)
	if err != nil {
		return 0, err
	}

	var rates map[string]map[string]float32
	err = json.Unmarshal(bodyBytes, &rates)
	if err != nil {
		return 0, err
	}

	return float64(rates[erp.coin][erp.currency]), nil
}

func readResponseBody(response *http.Response) ([]byte, error) {
	defer response.Body.Close()

	return ioutil.ReadAll(response.Body)
}
