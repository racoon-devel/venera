package webui

import (
	"fmt"
	"html/template"
	"time"
)

func tsToHumanReadable(ts int64) string {
	var tm = time.Duration(ts) * time.Second

	const day = 24 * time.Hour

	days := tm / day
	tm -= days * day

	hours := tm / time.Hour
	tm -= hours * time.Hour

	mins := tm / time.Minute
	tm -= mins * time.Minute

	secs := tm / time.Second

	return fmt.Sprintf("%d days, %d hours, %d min, %d sec",
		days, hours, mins, secs)
}

var (
	mainTpl = template.Must(template.New("main").Funcs(template.FuncMap{"ts": tsToHumanReadable}).Parse(`
	<html>
	<head></head>
	<body>
	<h2>Venera</h2>
	{{range $i, $t := $.Providers}}
		<a href="/task/new/{{$t}}">Новый поиск [ {{$t}} ]</a><br>
	{{- end}}
	</body>
	</html>
	`))
)
