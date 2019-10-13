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
	<head>
		<link rel="stylesheet" href="/ui/styles/default.css">
	</head>
	<body>
	<div id="header-wrapper">
		<div id="header" class="container">
			<div id="logo">
				<h1><a href="#">Venera</a></h1>
			</div>
			<div id="menu">
				<ul>
					<li class="current_page_item"><a href="/" accesskey="1" title="">Задачи</a></li>
					{{range $i, $t := $.Providers}}
					<li><a href="/task/new/{{$t}}" accesskey="{{$i}}">Новый поиск [ {{$t}} ]</a></li>
					{{- end}}
				</ul>
			</div>
		</div>
	</div>

	<div id="header-featured">
	</div>

	<div id="banner-wrapper">
		<div id="banner" class="container">
			<h2>Если у вас одновременно есть свобода и любовь</h2>
			<span>вам больше ничего не нужно. У вас все есть - то, ради чего дана жизнь. (с) </span>
		</div>
	</div>
	
	<div id="wrapper">
		<div id="page" class="container">
			<div id="content">
				<table>
				{{range $i, $t := $.Tasks}}
					<hr>
					<a href="/task/{{$t.ID}}">Task #{{$t.ID}} [ {{$t.Provider}} ]</a><br>
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
				</table>
			</div>
		</div>
	</div>
	
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
