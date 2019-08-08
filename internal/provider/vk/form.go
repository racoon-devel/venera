package vk

import (
	"html/template"
	"net/http"
)

var (
	formTpl = template.Must(template.New("vk").Parse(`
	<html>
	<head></head>
	<body>
	<h2>New task</h2>
	<form action="/task/new/vk" method="POST">
		<input type="submit" value="Submit">
	</form>
	</body>
	</html>
	`))
)

// NewTaskPageHandler - show search parameters form
func (ctx *VkProvider) NewTaskPageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {

	} else {
		formTpl.Execute(w, nil)
	}
}
