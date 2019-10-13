package webui

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"strings"

	//"racoondev.tk/gitea/racoon/venera/internal/dispatcher"
	"racoondev.tk/gitea/racoon/venera/internal/provider"
	"racoondev.tk/gitea/racoon/venera/internal/utils"
)

var Templates *template.Template

type mainContext struct {
	Providers []string
	//Tasks     []dispatcher.TaskInfo
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

	Templates, err = root.ParseFiles(tmplFiles...)
	if err != nil {
		return err
	}

	return nil
}

// MainPageHandler - show main admin page
func MainPageHandler(w http.ResponseWriter, r *http.Request) {
	var ctx mainContext
	ctx.Providers = provider.GetAvailable()
	// ctx.Tasks = dispatcher.Describe()
	fmt.Println(Templates.DefinedTemplates())
	Templates.ExecuteTemplate(w, "main", nil)
}

func ShowError(w http.ResponseWriter, err error) {
	errorTpl.Execute(w, err.Error())
}
