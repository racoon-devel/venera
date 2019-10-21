package webui

import (
	"fmt"
	"html/template"
	"time"

	"racoondev.tk/gitea/racoon/venera/internal/utils"

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

func StatusToHumanReadable(status types.SessionStatus) string {
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

func mod2(n int) int {
	return n % 2
}

func inc(n uint) uint {
	return n + 1
}

func hightlight(text string, matches []utils.TextMatch) template.HTML {
	return template.HTML(utils.Highlight(text, matches, "<mark>", "</mark>"))
}
