package webui

import (
	"net/http"

	"racoondev.tk/gitea/racoon/venera/internal/dispatcher"
	"racoondev.tk/gitea/racoon/venera/internal/provider"
)

type mainContext struct {
	Providers []string
	Tasks     []dispatcher.TaskInfo
}

// MainPageHandler - show main admin page
func MainPageHandler(w http.ResponseWriter, r *http.Request) {
	var ctx mainContext
	ctx.Providers = provider.GetAvailable()
	ctx.Tasks = dispatcher.Describe()
	mainTpl.Execute(w, &ctx)
}

func ShowError(w http.ResponseWriter, err error) {
	errorTpl.Execute(w, err.Error())
}
