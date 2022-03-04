package web_api

import (
	hederaHelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/hedera"
	httpHelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/http"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strings"
)

var (
	getHbarPriceHeaders = map[string]string{"Accepts": "application/json"}
)

type Client struct {
	apiCfg               config.HederaWebApi
	fullHederaGetHbarUrl string
	httpClient           *http.Client
	logger               *log.Entry
}

func NewClient(apiCfg config.HederaWebApi) *Client {
	return &Client{
		apiCfg:               apiCfg,
		httpClient:           new(http.Client),
		fullHederaGetHbarUrl: strings.Join([]string{apiCfg.BaseUrl, apiCfg.Endpoints.GetHbarPriceInUsd}, "/"),
		logger:               config.GetLoggerFor("Hedera Web API Client"),
	}
}

func (c *Client) GetHBARUsdPrice() (price decimal.Decimal, err error) {

	responseBody, err := httpHelper.Get(c.httpClient, c.fullHederaGetHbarUrl, getHbarPriceHeaders, c.logger)
	if err != nil {
		return decimal.Decimal{}, err
	}

	hederaFileRate, err := hederaHelper.GetHederaFileRateFromResponseBody(responseBody, c.logger)
	if err == nil {
		price = hederaFileRate.CurrentRate
	}

	return price, err
}
