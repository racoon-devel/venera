package mamba

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"racoondev.tk/gitea/racoon/venera/internal/utils"
)

func parseForm(r *http.Request, editMode bool) (*searchSettings, error) {
	err := r.ParseForm()
	if err != nil {
		return nil, err
	}

	ctx := searchSettings{}

	if len(r.Form["likes"]) != 1 || len(r.Form["likes"][0]) == 0 {
		return nil, fmt.Errorf("Field 'likes' must be not empty")
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
		return nil, fmt.Errorf("Field 'ageFrom' must be not empty")
	}

	if len(r.Form["ageTo"]) != 1 || len(r.Form["ageTo"][0]) == 0 {
		return nil, fmt.Errorf("Field 'ageTo' must be not empty")
	}

	if len(r.Form["city"]) != 1 || len(r.Form["city"][0]) == 0 {
		return nil, fmt.Errorf("Field 'city' must be not empty")
	}

	if len(r.Form["rater"]) != 1 || len(r.Form["rater"][0]) == 0 {
		return nil, fmt.Errorf("Field 'rater' must be not empty")
	}

	u, err := strconv.ParseUint(r.Form["ageFrom"][0], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("Field 'ageFrom' must be integer")
	}

	ctx.AgeFrom = uint(u)

	u, err = strconv.ParseUint(r.Form["ageTo"][0], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("Field 'ageTo' must be integer")
	}

	ctx.AgeTo = uint(u)

	ctx.City = r.Form["city"][0]
	ctx.Rater = r.Form["rater"][0]

	return &ctx, nil
}
