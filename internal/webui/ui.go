package webui

import (
	"net/http"

	"racoondev.tk/gitea/racoon/venera/internal/provider"
)

type mainContext struct {
	Providers []string
}

// MainPageHandler - show main admin page
func MainPageHandler(w http.ResponseWriter, r *http.Request) {
	var ctx mainContext
	ctx.Providers = provider.GetAvailable()
	mainTpl.Execute(w, &ctx)
}
