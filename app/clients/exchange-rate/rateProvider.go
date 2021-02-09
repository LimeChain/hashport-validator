package exchangerate

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type ExchangeRateProvider struct {
	httpClient *http.Client
	rateURL    string
	coin       string
	currency   string
	retries    int
	rate       float64
}

func NewExchangeRateProvider(coin string, currency string) *ExchangeRateProvider {
	return &ExchangeRateProvider{
		httpClient: &http.Client{},
		rateURL:    fmt.Sprintf("https://api.coingecko.com/api/v3/simple/price?ids=%s&vs_currencies=%s", coin, currency),
		coin:       coin,
		currency:   currency,
		retries:    0,
	}
}

func (erp *ExchangeRateProvider) Monitor() {
	go erp.getRateFromGecko()
}

func (erp *ExchangeRateProvider) GetRate() (float64, error) {
	if erp.rate > 0 {
		return erp.rate, nil
	}

	return 0, errors.New(fmt.Sprintf("Could not retrieve exchange rate for [%s] against [%s]: Rate is not retrieved yet", erp.coin, erp.currency))
}

func (erp *ExchangeRateProvider) getRateFromGecko() {
	response, err := erp.httpClient.Get(erp.rateURL)
	if err != nil {
		erp.retry(err)
		return
	}

	bodyBytes, err := readResponseBody(response)
	if err != nil {
		erp.retry(err)
		return
	}

	var rates map[string]map[string]float32
	err = json.Unmarshal(bodyBytes, &rates)
	if err != nil {
		erp.retry(err)
		return
	}

	erp.rate = float64(rates[erp.coin][erp.currency])

	time.Sleep(1 * time.Hour)
	erp.getRateFromGecko()
}

func (erp *ExchangeRateProvider) retry(err error) {
	erp.retries++
	fmt.Errorf("Could not retrieve exchange rate for [%s] against [%s]: %s", erp.coin, erp.currency, err)
	if erp.retries < 10 {
		erp.getRateFromGecko()
	}
}

func readResponseBody(response *http.Response) ([]byte, error) {
	defer response.Body.Close()

	return ioutil.ReadAll(response.Body)
}
