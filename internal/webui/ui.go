package webui

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"strings"

	"racoondev.tk/gitea/racoon/venera/internal/types"
	"racoondev.tk/gitea/racoon/venera/internal/utils"
)

var templates *template.Template

type ItemContext struct {
	*types.PersonRecord
	Actions []types.Action
}

type ResultContext struct {
	Pages      []uint
	Page       uint
	Total      uint
	Tasks      []types.TaskRecord
	Results    []*ItemContext
	TaskFilter uint
	Ascending  bool
	ViewMode   uint
	Favourite  bool
	Rating     uint
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

	root := template.New("root").Funcs(template.FuncMap{
		"ts":        tsToHumanReadable,
		"status":    StatusToHumanReadable,
		"mod2":      mod2,
		"inc":       inc,
		"highlight": hightlight,
	})

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

func DisplayEditTask(w http.ResponseWriter, provider string, context interface{}) {
	fmt.Println(templates.DefinedTemplates())
	templates.ExecuteTemplate(w, "edit."+provider, context)
}

func DisplayResults(w http.ResponseWriter, results *ResultContext) {
	templates.ExecuteTemplate(w, "results", results)
}

func DisplayResult(w http.ResponseWriter, result *ItemContext) {
	templates.ExecuteTemplate(w, "result", result)
}
