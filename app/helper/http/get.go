package http

import (
	"encoding/json"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func Get(client client.HttpClient, url string, headers map[string]string, responseStruct interface{}, log *log.Entry) (err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Errorf("Error while creating http request struct. Error: [%v]", err)
		return err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("Error while sending request to server. Error: [%v]", err)
		return err
	}

	err = json.NewDecoder(resp.Body).Decode(responseStruct)

	return err
}
