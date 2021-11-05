package tinder

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/racoon-devel/venera/internal/utils"
)

func parseForm(r *http.Request, editMode bool) (*searchSettings, *tinderAuth, error) {
	err := r.ParseForm()
	if err != nil {
		return nil, nil, err
	}

	ctx := searchSettings{}
	var auth *tinderAuth

	if !editMode {

		if len(r.Form["tel"]) != 1 || len(r.Form["tel"][0]) == 0 {
			return nil, nil, fmt.Errorf("Field 'tel' must be not empty")
		}
		ctx.Tel = r.Form["tel"][0]

		auth = newTinderAuth(ctx.Tel)

		//if len(r.Form["login_token"]) != 1 || len(r.Form["login_token"][0]) == 0 {
		//	return nil, nil, fmt.Errorf("You must fill login code")
		//}

		//auth.LoginToken = r.Form["login_token"][0]

		if len(r.Form["code"]) != 1 || len(r.Form["code"][0]) == 0 {
			return nil, nil, fmt.Errorf("You must fill login code")
		}

		auth.LoginCode = r.Form["code"][0]
	}

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

	if len(r.Form["longitude"]) != 1 || len(r.Form["longitude"][0]) == 0 {
		return nil, nil, fmt.Errorf("Field 'longitude' must be not empty")
	}

	if len(r.Form["latitude"]) != 1 || len(r.Form["latitude"][0]) == 0 {
		return nil, nil, fmt.Errorf("Field 'latitude' must be not empty")
	}

	if len(r.Form["rater"]) != 1 || len(r.Form["rater"][0]) == 0 {
		return nil, nil, fmt.Errorf("Field 'rater' must be not empty")
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

	f, err := strconv.ParseFloat(r.Form["latitude"][0], 32)
	if err != nil {
		return nil, nil, fmt.Errorf("Field 'latitude' must be float")
	}

	ctx.Latitude = float32(f)

	f, err = strconv.ParseFloat(r.Form["longitude"][0], 32)
	if err != nil {
		return nil, nil, fmt.Errorf("Field 'Longitude' must be float")
	}

	ctx.Longitude = float32(f)

	ctx.Rater = r.Form["rater"][0]

	return &ctx, auth, nil
}
