package badoo

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

	if !editMode {

		if len(r.Form["email"]) != 1 || len(r.Form["email"][0]) == 0 {
			return nil, fmt.Errorf("Field 'email' must be not empty")
		}
		ctx.Email = r.Form["email"][0]

		if len(r.Form["password"]) != 1 || len(r.Form["password"][0]) == 0 {
			return nil, fmt.Errorf("You must fill login password")
		}

		ctx.Password = r.Form["password"][0]

		if len(r.Form["rater"]) != 1 || len(r.Form["rater"][0]) == 0 {
			return nil, fmt.Errorf("Field 'rater' must be not empty")
		}

		ctx.Rater = r.Form["rater"][0]
	}

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

	if len(r.Form["longitude"]) != 1 || len(r.Form["longitude"][0]) == 0 {
		return nil, fmt.Errorf("Field 'longitude' must be not empty")
	}

	if len(r.Form["latitude"]) != 1 || len(r.Form["latitude"][0]) == 0 {
		return nil, fmt.Errorf("Field 'latitude' must be not empty")
	}

	f, err := strconv.ParseFloat(r.Form["latitude"][0], 32)
	if err != nil {
		return nil, fmt.Errorf("Field 'latitude' must be float")
	}

	ctx.Latitude = float32(f)

	f, err = strconv.ParseFloat(r.Form["longitude"][0], 32)
	if err != nil {
		return nil, fmt.Errorf("Field 'Longitude' must be float")
	}

	ctx.Longitude = float32(f)

	return &ctx, nil
}
