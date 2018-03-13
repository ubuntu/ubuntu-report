package sender

import (
	"bytes"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const URL = "https://metrics.ubuntu.com"

// Send to url the json data
func Send(url string, data []byte) error {
	log.Debugf("sending %s to %s", data, url)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return errors.Wrap(err, "couldn't create http request")
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, "couldn't send post http request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("incorrect status code received: %s", resp.Status)
	}

	_, err = ioutil.ReadAll(resp.Body)
	return errors.Wrap(err, "POST body answer contained an error")
}
