package export

import (
	"fmt"
	"net/http"
	"strconv"
)

func parseForm(r *http.Request) (*searchSettings, error) {
	err := r.ParseForm()
	if err != nil {
		return nil, err
	}

	ctx := searchSettings{}

	if len(r.Form["task"]) > 0 && len(r.Form["task"][0]) != 0 {
		u, err := strconv.ParseUint(r.Form["task"][0], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("Field 'task' must be integer")
		}
		ctx.taskID = uint(u)
	}

	ctx.ExportFavourite = len(r.Form["favourite"]) > 0 && len(r.Form["favourite"][0]) != 0
	ctx.ExportAbout = len(r.Form["about"]) > 0 && len(r.Form["about"][0]) != 0
	ctx.ExportPhotos = len(r.Form["photos"]) > 0 && len(r.Form["photos"][0]) != 0
	ctx.ExportDump = len(r.Form["dump"]) > 0 && len(r.Form["dump"][0]) != 0

	fmt.Printf("%+v\n", &ctx)

	return &ctx, nil
}
