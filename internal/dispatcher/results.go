package dispatcher

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/gorilla/mux"
	"racoondev.tk/gitea/racoon/venera/internal/types"
	"racoondev.tk/gitea/racoon/venera/internal/webui"
)

const resultsPerPage uint = 5

func getField(query url.Values, name string) uint {
	param, ok := query[name]
	if ok && len(param) != 0 {
		val, err := strconv.ParseUint(param[0], 10, 32)
		if err == nil {
			return uint(val)
		}
	}

	return 0
}

func resultsHandler(w http.ResponseWriter, r *http.Request) {
	var res webui.ResultContext

	res.TaskFilter = getField(r.URL.Query(), "task")
	res.Page = getField(r.URL.Query(), "page")
	res.ViewMode = getField(r.URL.Query(), "mode")

	orderParam, ok := r.URL.Query()["order"]
	if ok && len(orderParam) != 0 {
		if orderParam[0] == "asc" {
			res.Ascending = true
		}
	}

	res.Tasks = dispatcher.db.LoadTasks()

	var err error
	var persons []types.PersonRecord
	persons, res.Total, err = dispatcher.db.LoadPersons(res.TaskFilter, res.Ascending, resultsPerPage, res.Page*resultsPerPage)
	if err != nil {
		webui.DisplayError(w, err)
		return
	}

	res.Results = make([]*webui.ItemContext, len(persons))
	for i := 0; i < len(persons); i++ {
		item := webui.ItemContext{}
		item.PersonRecord = &persons[i]
		wrapItem(&item)
		res.Results[i] = &item
	}

	pages := res.Total / resultsPerPage
	if res.Total%resultsPerPage != 0 {
		pages++
	}

	res.Pages = make([]uint, pages)
	for i := uint(0); i < pages; i++ {
		res.Pages[i] = i
	}

	webui.DisplayResults(w, &res)
}

func wrapItem(item *webui.ItemContext) {
	prov := getTaskProvider(item.TaskID)
	if prov == nil {
		return
	}

	item.Actions = prov.GetResultActions(item.PersonRecord)
}

func resultHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["result"], 10, 32)
	if err != nil {
		webui.DisplayError(w, err)
		return
	}

	result, err := dispatcher.db.LoadPerson(uint(id))
	if err != nil {
		webui.DisplayError(w, err)
		return
	}

	item := webui.ItemContext{}
	item.PersonRecord = result
	wrapItem(&item)

	webui.DisplayResult(w, &item)
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["result"], 10, 32)
	if err != nil {
		webui.DisplayError(w, err)
		return
	}

	dispatcher.db.DeletePerson(uint(id))
	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
}
