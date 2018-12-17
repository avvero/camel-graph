package model

import (
	"net/http"
	"io/ioutil"
	"time"
	"errors"
)

func callEndpoint(url string, auth *Authorization) ([]byte, error) {
	client := &http.Client{
		Timeout: time.Duration(60 * time.Second),
	}
	req, err := http.NewRequest("GET", url, nil)
	if auth != nil {
		req.SetBasicAuth(auth.Login, auth.Pass)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, errors.New("Status " + resp.Status)
	}
	defer resp.Body.Close()
	//return json.NewDecoder(resp.Body).Decode(target)
	return ioutil.ReadAll(resp.Body)
}
