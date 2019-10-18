package tinder

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

const (
	userAgent = "Mozilla/5.0 (iPhone; CPU iPhone OS 11_2_5 like Mac OS X) AppleWebKit/604.5.6 (KHTML, like Gecko) Mobile/15D60 AKiOSSDK/4.29.0"
)

type tinderAuth struct {
	Tel        string
	LoginToken string
	LoginCode  string
	APIToken   string
}

func tinderRequest(url string, requestBody io.Reader) ([]byte, error) {
	request, err := http.NewRequest("POST", url, requestBody)
	if err != nil {
		return nil, err
	}

	request.Header.Add("User-Agent", userAgent)

	if requestBody != nil {
		request.Header.Add("Content-Type", "application/json")
	}
	client := http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func newTinderAuth(tel string) *tinderAuth {
	return &tinderAuth{Tel: tel}
}

func (self *tinderAuth) RequestCode() error {
	type codeResponse struct {
		LoginToken string `json:"login_request_code"`
	}

	self.LoginCode = ""
	self.LoginToken = ""
	self.APIToken = ""

	url := "https://graph.accountkit.com/v1.2/start_login?access_token=AA%7C464891386855067%7Cd1891abb4b0bcdfa0580d9b839f4a522&credentials_type=phone_number&fb_app_events_enabled=1&fields=privacy_policy%2Cterms_of_service&locale=fr_FR&phone_number=" +
		url.QueryEscape(self.Tel) +
		"&response_type=token&sdk=ios"

	body, err := tinderRequest(url, nil)
	if err != nil {
		return err
	}

	cr := codeResponse{}
	if err := json.Unmarshal(body, &cr); err != nil {
		return err
	}

	self.LoginToken = cr.LoginToken

	return nil
}

func (self *tinderAuth) RequestToken() error {
	self.APIToken = ""

	url := "https://graph.accountkit.com/v1.2/confirm_login?access_token=AA%7C464891386855067%7Cd1891abb4b0bcdfa0580d9b839f4a522&confirmation_code=" +
		self.LoginCode + "&credentials_type=phone_number&fb_app_events_enabled=1&fields=privacy_policy%2Cterms_of_service&locale=fr_FR&login_request_code=" +
		self.LoginToken + "&phone_number=" + self.Tel + "&response_type=token&sdk=ios"

	body, err := tinderRequest(url, nil)
	if err != nil {
		return err
	}

	type verifyResponse struct {
		AccessToken string `json:"access_token"`
		AccessID    string `json:"id"`
	}

	type getTokenRequest struct {
		Token  string `json:"token"`
		ID     string `json:"id"`
		Client string `json:"client_version"`
	}

	verify := &verifyResponse{}
	err = json.Unmarshal(body, verify)
	if err != nil {
		return err
	}

	req := &getTokenRequest{Token: verify.AccessToken, ID: verify.AccessID, Client: "9.0.1"}

	requestBody, err := json.Marshal(req)
	if err != nil {
		return err
	}

	url = "https://api.gotinder.com/v2/auth/login/accountkit"

	body, err = tinderRequest(url, bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}

	// лень делать преобразование типов
	type data struct {
		APIToken string `json:"api_token"`
	}

	type response struct {
		Data data
	}

	tokenHolder := &response{}

	err = json.Unmarshal(body, tokenHolder)
	if err != nil {
		return err
	}

	self.APIToken = tokenHolder.Data.APIToken

	return nil
}
