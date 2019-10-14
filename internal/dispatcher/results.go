package dispatcher

import (
	"net/http"
	"strconv"

	"racoondev.tk/gitea/racoon/venera/internal/webui"
)

func resultsHandler(w http.ResponseWriter, r *http.Request) {
	var res webui.ResultContext

	taskParam, ok := r.URL.Query()["task"]
	if ok && len(taskParam) != 0 {
		ID, err := strconv.ParseUint(taskParam[0], 10, 32)
		if err == nil {
			res.TaskFilter = uint(ID)
		}
	}

	orderParam, ok := r.URL.Query()["order"]
	if ok && len(orderParam) != 0 {
		if orderParam[0] == "asc" {
			res.Ascending = true
		}
	}

	res.Tasks = dispatcher.db.LoadTasks()

	var err error
	res.Results, err = dispatcher.db.LoadPersons(res.TaskFilter, res.Ascending)
	if err != nil {
		webui.DisplayError(w, err)
		return
	}

	webui.DisplayResults(w, &res)
}
