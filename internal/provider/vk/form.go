package vk

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/racoon-devel/venera/internal/utils"
)

func parseForm(r *http.Request, editMode bool) (*searchSettings, error) {
	err := r.ParseForm()
	if err != nil {
		return nil, err
	}

	ctx := searchSettings{}

	if !editMode {

		if len(r.Form["login"]) != 1 || len(r.Form["login"][0]) == 0 {
			return nil, fmt.Errorf("field 'login' must be not empty")
		}
		ctx.Login = r.Form["login"][0]

		if len(r.Form["password"]) != 1 || len(r.Form["password"][0]) == 0 {
			return nil, fmt.Errorf("you must fill login password")
		}
		ctx.Password = r.Form["password"][0]

		if len(r.Form["rater"]) != 1 || len(r.Form["rater"][0]) == 0 {
			return nil, fmt.Errorf("field 'rater' must be not empty")
		}
		ctx.Rater = r.Form["rater"][0]

		if len(r.Form["keywords"]) != 1 || len(r.Form["keywords"][0]) == 0 {
			return nil, fmt.Errorf("field 'keywords' must be not empty")
		}
		ctx.Keywords = strings.Split(r.Form["keywords"][0], ",")
		utils.TrimArray(ctx.Keywords)
	}

	if len(r.Form["likes"]) != 1 || len(r.Form["likes"][0]) == 0 {
		return nil, fmt.Errorf("field 'likes' must be not empty")
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
		return nil, fmt.Errorf("field 'ageFrom' must be not empty")
	}

	if len(r.Form["ageTo"]) != 1 || len(r.Form["ageTo"][0]) == 0 {
		return nil, fmt.Errorf("Field 'ageTo' must be not empty")
	}

	if len(r.Form["city"]) != 1 || len(r.Form["city"][0]) == 0 {
		return nil, fmt.Errorf("field 'city' must be not empty")
	}
	ctx.City = r.Form["city"][0]

	u, err := strconv.ParseUint(r.Form["ageFrom"][0], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("Field 'ageFrom' must be integer")
	}

	ctx.AgeFrom = uint(u)

	u, err = strconv.ParseUint(r.Form["ageTo"][0], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("field 'ageTo' must be integer")
	}

	ctx.AgeTo = uint(u)

	return &ctx, nil
}
