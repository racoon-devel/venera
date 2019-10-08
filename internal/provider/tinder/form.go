package tinder

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/ccding/go-logging/logging"

	"racoondev.tk/gitea/racoon/venera/internal/types"
	"racoondev.tk/gitea/racoon/venera/internal/utils"
)

var (
	formTpl = template.Must(template.New("tinder").Parse(`
	<html>
	<head></head>
	<body>
	<h2>New task</h2>
	<form action="/task/new/tinder" method="POST">
		User:<input type="text" name="user"><br>
		Password:<input type="password" name="password"><br>
		Token:<input type="token" name="token"><br>
		<a href="https://www.facebook.com/v2.6/dialog/oauth?redirect_uri=fb464891386855067%3A%2F%2Fauthorize%2F&scope=user_birthday,user_photos,user_education_history,email,user_relationship_details,user_friends,user_work_history,user_likes&response_type=token%2Csigned_request&client_id=464891386855067">Get Token</a><br>
		Age:<input type="text" name="ageFrom"> - <input type="text" name="ageTo"><br>
		Likes:<input type="text" name="likes"><br>
		Dislikes:<input type="text" name="dislikes"><br>
		<input type="submit" value="Submit">
	</form>
	</body>
	</html>
	`))
)

func getSearchSettings(r *http.Request) (*searchSettings, error) {
	err := r.ParseForm()
	if err != nil {
		return nil, err
	}

	ctx := searchSettings{}

	if len(r.Form["user"]) != 1 || len(r.Form["user"][0]) == 0 {
		return nil, fmt.Errorf("Field 'user' must be not empty")
	}
	ctx.User = r.Form["user"][0]

	if len(r.Form["password"]) != 1 || len(r.Form["password"][0]) == 0 {
		return nil, fmt.Errorf("Field 'password' must be not empty")
	}
	ctx.Password = r.Form["password"][0]

	if len(r.Form["token"]) != 1 || len(r.Form["token"][0]) == 0 {
		return nil, fmt.Errorf("Field 'token' must be not empty")
	}
	ctx.Token = r.Form["token"][0]

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

	if ctx.AgeFrom > ctx.AgeTo {
		return nil, fmt.Errorf("Field 'ageTo' must be greater or equal than field 'ageFrom'")
	}

	fmt.Println(ctx)

	return &ctx, nil
}

// ShowSearchPage - show search parameters form
func (ctx *TinderProvider) ShowSearchPage(w http.ResponseWriter) {
	formTpl.Execute(w, nil)
}

func (ctx *TinderProvider) GetSearchSession(log *logging.Logger, r *http.Request) (types.SearchSession, error) {
	settings, err := getSearchSettings(r)
	if err != nil {
		return nil, err
	}

	return NewSession(settings, log), nil
}
