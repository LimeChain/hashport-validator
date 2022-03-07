package http

import (
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
)

func Get(client client.HttpClient, url string, headers map[string]string, log *log.Entry) (responseBodyAsBytes []byte, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Errorf("Error while creating http request struct. Error: [%v]", err)
		return responseBodyAsBytes, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("Error while sending request to server. Error: [%v]", err)
		return responseBodyAsBytes, err
	}

	responseBodyAsBytes, _ = ioutil.ReadAll(resp.Body)
	return responseBodyAsBytes, nil
}
