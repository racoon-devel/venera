package vk

import (
	"errors"
	"fmt"
	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/racoon-devel/venera/internal/navigator"
	"github.com/racoon-devel/venera/internal/utils"
	"net/url"
	"regexp"
	"time"
)

func (session *searchSession) signIn() error {
	u := fmt.Sprintf("https://oauth.vk.com/authorize?client_id=%d&display=popup&redirect_uri=http://racoondev.tk/&scope=friends,photos&response_type=token&v=%s&state=123456&revoke=1", appId, api.Version)
	nav, err := navigator.Open(session.log, u)
	if err != nil {
		return err
	}
	defer nav.Close()

	nav.SetErrorReportsPath(utils.Configuration.Directories.Downloads)

	err = nav.Batch("Sign in VK").
		Type(`#login_submit > div > div > input:nth-child(8)`, session.state.Search.Login).
		Type(`#login_submit > div > div > input:nth-child(10)`, session.state.Search.Password).
		Sleep(2 * time.Second).
		Click(`#install_allow`).
		Sleep(2 * time.Second).
		Click(`//*[@id="oauth_wrap_content"]/div[3]/div/div[1]/button[1]`).
		Sleep(2 * time.Second).
		Error()
	if err != nil {
		return err
	}

	redirect, err := url.Parse(nav.Address())
	if err != nil {
		return err
	}

	session.log.Debugf("URL = ", nav.Address())

	authorizeURLs, ok := redirect.Query()["authorize_url"]
	if !ok {
		return errors.New("couldn't get authorize url")
	}

	authorizeURL, err := url.QueryUnescape(authorizeURLs[0])
	if err != nil {
		return err
	}

	completeURL, err := url.Parse(authorizeURL)
	if err != nil {
		return err
	}

	r := regexp.MustCompile(`access_token=([\w\d]+)`)
	matched := r.FindStringSubmatch(completeURL.Fragment)
	if len(matched) != 2 {
		return errors.New("couldn't extract access token")
	}

	session.mutex.Lock()
	session.state.AccessToken = matched[1]
	session.state.LastAuthTime = time.Now()
	session.api = api.NewVK(session.state.AccessToken)
	session.api.Limit = api.LimitUserToken
	session.mutex.Unlock()

	session.log.Debugf("VK access_token = %s", matched[1])

	return nil
}
