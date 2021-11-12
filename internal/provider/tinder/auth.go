package tinder

import (
	"errors"
	"strings"
	"time"

	"github.com/ccding/go-logging/logging"
	"github.com/racoon-devel/venera/internal/navigator"
	"github.com/racoon-devel/venera/internal/utils"
)

type tinderAuth struct {
	Login    string
	Password string
	APIToken string
}

func newTinderAuth(login, password string) *tinderAuth {
	return &tinderAuth{Login: login, Password: password}
}

func (a *tinderAuth) SignIn(log *logging.Logger) error {
	nav, err := navigator.Open(log, "https://accounts.google.com/signin")
	if err != nil {
		return err
	}
	defer nav.Close()

	nav.SetErrorReportsPath(utils.Configuration.Directories.Downloads)

	var greeting string
	err = nav.Batch("Sign in Google").
		Type(`//*[@id="identifierId"]`, a.Login).
		Click(`//*[@id="identifierNext"]/div/button`).
		Sleep(1*time.Second).
		Type(`//*[@id="password"]/div[1]/div/div[1]/input`, a.Password).
		Sleep(2*time.Second).
		Click(`//*[@id="passwordNext"]/div/button`).
		Fetch(`//*[@id="yDmH0d"]/c-wiz/div/div[2]/c-wiz/c-wiz/div/div[3]/div/div/header/h1`, &greeting).
		Error()

	if err != nil {
		return err
	}

	if !strings.HasPrefix(greeting, "Добро пожаловать") {
		return errors.New("login on Google failed")
	}

	headers := make(map[string]string)
	err = nav.Batch("Sign in Tinder").
		Goto("https://tinder.com").
		Click(`//div[1]/div/div[1]/div/main/div[1]/div/div/div/div/header/div/div[2]/div[2]/a`).
		Sleep(3 * time.Second).
		Click(`//div[2]/div/div/div[1]/div/div[3]/span/div[1]/div/button`).
		Sleep(5 * time.Second).
		CaptureHeaders(headers).
		Error()

	if err != nil {
		return nil
	}

	ok := false
	a.APIToken, ok = headers["x-auth-token"]
	if !ok {
		return errors.New("cannot extract api token")
	}

	return nil
}
