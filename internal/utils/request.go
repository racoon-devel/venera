package utils

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func HttpRequest(url string) (response []byte, err error) {
	client := &http.Client{
		Timeout: time.Duration(Configuration.Http.Timeout) * time.Second,
	}

	var resp *http.Response
	resp, err = client.Get(url)

	if err != nil {
		return
	}

	response, err = ioutil.ReadAll(resp.Body)

	return
}

func JsonRequest(url string, item interface{}) []byte {
	client := &http.Client{
		Timeout: time.Duration(Configuration.Http.Timeout) * time.Second,
	}

	body, err := json.Marshal(item)

	if err != nil {
		log.Println("Create JSON request failed:", err)
		return nil
	}

	req, err := http.NewRequest("GET", url, bytes.NewReader(body))

	if err != nil {
		log.Println("Create HTTP request failed:", err)
		return nil
	}

	resp, err := client.Do(req)

	if err != nil {
		log.Println("HTTP request failed:", err)
		return nil
	}

	if resp.StatusCode != 200 {
		log.Println("Invalid status code:", resp.StatusCode)
		return nil
	}

	response, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Println("Read response failed: ", response)
		return nil
	}

	return response
}
