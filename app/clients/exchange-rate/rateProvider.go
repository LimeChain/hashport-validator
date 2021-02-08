package exchange_rate

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type ExchangeRateProvider struct {
	httpClient *http.Client
	rateURL    string
}

func NewExchangeRateProvider() *ExchangeRateProvider {
	return &ExchangeRateProvider{
		httpClient: &http.Client{},
		rateURL:    "https://api.coingecko.com/api/v3/simple/price?ids=hedera-hashgraph&vs_currencies=eth",
	}
}

func (erp *ExchangeRateProvider) Monitor() {
	go erp.GetRate()
}

func (erp *ExchangeRateProvider) GetRate() {
	response, err := erp.httpClient.Get(erp.rateURL)
	if err != nil {
		// handle error properly;
	}

	bodyBytes, err := readResponseBody(response)
	if err != nil {
		// handle error properly;
	}

	var rates map[string]map[string]float32
	err = json.Unmarshal(bodyBytes, &rates)
	if err != nil {
		// handle error properly;
	}

	fmt.Println(rates["hedera-hashgraph"]["eth"])

	time.Sleep(5 * time.Second)
	erp.GetRate()
}

func readResponseBody(response *http.Response) ([]byte, error) {
	defer response.Body.Close()

	return ioutil.ReadAll(response.Body)
}
