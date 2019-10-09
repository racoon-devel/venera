package webui

import (
	"fmt"
	"html/template"
	"time"

	"racoondev.tk/gitea/racoon/venera/internal/types"
)

func tsToHumanReadable(ts time.Duration) string {
	var tm = time.Duration(ts)

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

func statusToHumanReadable(status types.SessionStatus) string {
	switch status {

	case types.StatusIdle:
		return "Idle"

	case types.StatusRunning:
		return "Running"

	case types.StatusStopped:
		return "Stopped"

	case types.StatusError:
		return "Error"

	default:
		return "Unknown"
	}
}

var (
	mainTpl = template.Must(template.New("main").Funcs(template.FuncMap{"ts": tsToHumanReadable, "status": statusToHumanReadable}).Parse(`
	<html>
	<head></head>
	<body>
	<h2>Venera</h2>
	<table>
	{{range $i, $t := $.Providers}}
		<a href="/task/new/{{$t}}">Новый поиск [ {{$t}} ]</a><br>
	{{- end}}
	{{range $i, $t := $.Tasks}}
	<hr>
		Task #{{$t.ID}} [ {{$t.Provider}} ]<br>
		<small>{{ts $t.Remaining }}</small><br>
		<small>Status: {{status $t.Status}}</small><br>
		<table><td>
		{{ if eq $t.Status 1 }}
		<a href="/task/pause/{{$t.ID}}">Pause</a>
		{{ else }}
		<a href="/task/run/{{$t.ID}}">Run</a>
		{{ end }}
		</td>
		<td><a href="/task/stop/{{$t.ID}}">Stop</a></td>
		<td><a href="/task/delete/{{$t.ID}}">Delete</a></td>
		</table>
	{{- end}}
	</body>
	</html>
	`))

	errorTpl = template.Must(template.New("error").Parse(`
	<html>
	<head></head>
	<body>
	<h2>Try again</h2>
	<font color="red">{{$}}</font>
	</body>
	</html>
	`))
)
