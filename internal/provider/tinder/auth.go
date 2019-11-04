package tinder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

const (
	userAgent = "Mozilla/5.0 (iPhone; CPU iPhone OS 11_2_5 like Mac OS X) AppleWebKit/604.5.6 (KHTML, like Gecko) Mobile/15D60 AKiOSSDK/4.29.0"
)

type tinderAuth struct {
	Tel          string
	RefreshToken string
	LoginCode    string
	APIToken     string
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
		Meta struct {
			Status int
		}
		Data struct {
			SmsSent bool `json:"sms_sent"`
		}
	}

	self.LoginCode = ""
	self.RefreshToken = ""
	self.APIToken = ""

	url := "https://api.gotinder.com/v2/auth/sms/send?auth_type=sms&locale=ru"
	buffer := bytes.NewBuffer([]byte(fmt.Sprintf(`{"phone_number":"%s"}`, self.Tel)))

	body, err := tinderRequest(url, buffer)
	if err != nil {
		return err
	}

	cr := codeResponse{}
	if err := json.Unmarshal(body, &cr); err != nil {
		return err
	}

	if cr.Meta.Status != 200 || !cr.Data.SmsSent {
		return fmt.Errorf("invalid server response: %+v", &cr)
	}

	return nil
}

func (self *tinderAuth) ValidateCode(code string) error {
	type validateRequest struct {
		Code     string `json:"otp_code"`
		Tel      string `json:"phone_number"`
		IsUpdate bool   `json:"is_update"`
	}

	type validateResponse struct {
		Meta struct {
			Status int
		}
		Data struct {
			RefreshToken string `json:"refresh_token"`
			Validated    bool
		}
	}

	self.APIToken = ""
	self.LoginCode = code

	url := "https://api.gotinder.com/v2/auth/sms/validate?auth_type=sms&locale=ru"

	req := validateRequest{
		Code: self.LoginCode,
		Tel:  self.Tel,
	}

	requestBody, err := json.Marshal(&req)
	if err != nil {
		return err
	}

	body, err := tinderRequest(url, bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}

	verify := &validateResponse{}
	err = json.Unmarshal(body, verify)
	if err != nil {
		return err
	}

	if !verify.Data.Validated || verify.Meta.Status != 200 {
		return fmt.Errorf("invalid server response: %+v", verify)
	}

	self.RefreshToken = verify.Data.RefreshToken

	return nil
}

func (self *tinderAuth) Login() error {
	type loginRequest struct {
		Tel          string `json:"phone_number"`
		RefreshToken string `json:"refresh_token"`
	}

	type loginResponse struct {
		Meta struct {
			Status int
		}
		Data struct {
			RefreshToken string `json:"refresh_token"`
			APIToken     string `json:"api_token"`
			IsNewUser    bool   `json:"is_new_user"`
		}
	}

	self.APIToken = ""

	url := "https://api.gotinder.com/v2/auth/login/sms?locale=en"
	req := loginRequest{
		Tel:          self.Tel,
		RefreshToken: self.RefreshToken,
	}

	requestBody, err := json.Marshal(&req)
	if err != nil {
		return err
	}

	body, err := tinderRequest(url, bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}

	login := &loginResponse{}
	err = json.Unmarshal(body, login)
	if err != nil {
		return err
	}

	if login.Data.IsNewUser || login.Meta.Status != 200 {
		return fmt.Errorf("invalid server response: %+v", login)
	}

	self.RefreshToken = login.Data.RefreshToken
	self.APIToken = login.Data.APIToken

	return nil
}
