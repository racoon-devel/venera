package tinder

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/ccding/go-logging/logging"

	"racoondev.tk/gitea/racoon/venera/internal/types"
	"racoondev.tk/gitea/racoon/venera/internal/utils"
)

func getSearchSettings(r *http.Request) (*searchSettings, *tinderAuth, error) {
	err := r.ParseForm()
	if err != nil {
		return nil, nil, err
	}

	ctx := searchSettings{}

	if len(r.Form["tel"]) != 1 || len(r.Form["tel"][0]) == 0 {
		return nil, nil, fmt.Errorf("Field 'tel' must be not empty")
	}
	ctx.Tel = r.Form["tel"][0]

	auth := newTinderAuth(ctx.Tel)

	if len(r.Form["login_token"]) != 1 || len(r.Form["login_token"][0]) == 0 {
		return nil, nil, fmt.Errorf("You must fill login code")
	}

	auth.LoginToken = r.Form["login_token"][0]

	if len(r.Form["code"]) != 1 || len(r.Form["code"][0]) == 0 {
		return nil, nil, fmt.Errorf("You must fill login code")
	}

	auth.LoginCode = r.Form["code"][0]

	if len(r.Form["likes"]) != 1 || len(r.Form["likes"][0]) == 0 {
		return nil, nil, fmt.Errorf("Field 'likes' must be not empty")
	}
	ctx.Likes = strings.Split(r.Form["likes"][0], ",")
	utils.TrimArray(ctx.Likes)

	if len(r.Form["dislikes"]) == 1 || len(r.Form["dislikes"][0]) == 0 {
		ctx.Dislikes = strings.Split(r.Form["dislikes"][0], ",")
		utils.TrimArray(ctx.Dislikes)
	} else {
		ctx.Dislikes = make([]string, 0)
	}

	if len(r.Form["ageFrom"]) != 1 || len(r.Form["ageFrom"][0]) == 0 {
		return nil, nil, fmt.Errorf("Field 'ageFrom' must be not empty")
	}

	if len(r.Form["ageTo"]) != 1 || len(r.Form["ageTo"][0]) == 0 {
		return nil, nil, fmt.Errorf("Field 'ageTo' must be not empty")
	}

	u, err := strconv.ParseUint(r.Form["ageFrom"][0], 10, 64)
	if err != nil {
		return nil, nil, fmt.Errorf("Field 'ageFrom' must be integer")
	}

	ctx.AgeFrom = uint(u)

	u, err = strconv.ParseUint(r.Form["ageTo"][0], 10, 64)
	if err != nil {
		return nil, nil, fmt.Errorf("Field 'ageTo' must be integer")
	}

	ctx.AgeTo = uint(u)

	return &ctx, auth, nil
}

func (ctx *TinderProvider) GetSearchSession(log *logging.Logger, r *http.Request) (types.SearchSession, error) {
	settings, auth, err := getSearchSettings(r)
	if err != nil {
		return nil, err
	}

	if err := settings.SearchSettings.Validate(); err != nil {
		return nil, err
	}

	if err := auth.RequestToken(); err != nil {
		return nil, fmt.Errorf("Tinder auth failed: %+v", err)
	}

	settings.APIToken = auth.APIToken

	return NewSession(settings, log), nil
}
