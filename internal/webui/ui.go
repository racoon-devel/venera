package webui

import (
	"html/template"
	"io/ioutil"
	"net/http"
	"strings"

	"racoondev.tk/gitea/racoon/venera/internal/utils"
)

var templates *template.Template

type UIContext struct {
	PageSelected int
}

func LoadTemplates() error {
	var tmplFiles []string
	dir := utils.Configuration.Other.Content + "/templates"
	files, err := ioutil.ReadDir(utils.Configuration.Other.Content + "/templates")
	if err != nil {
		return err
	}
	for _, file := range files {
		filename := file.Name()
		if strings.HasSuffix(filename, ".html") {
			tmplFiles = append(tmplFiles, dir+"/"+filename)
		}
	}

	root := template.New("root").Funcs(template.FuncMap{"ts": tsToHumanReadable, "status": statusToHumanReadable})

	templates, err = root.ParseFiles(tmplFiles...)
	if err != nil {
		return err
	}

	return nil
}

func DisplayMain(w http.ResponseWriter, context interface{}) {
	templates.ExecuteTemplate(w, "main", context)
}

func DisplayError(w http.ResponseWriter, err error) {
	templates.ExecuteTemplate(w, "error", err)
}

func DisplayNewTask(w http.ResponseWriter, provider string) {
	templates.ExecuteTemplate(w, "new."+provider, nil)
}
