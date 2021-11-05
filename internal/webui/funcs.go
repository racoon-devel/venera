package webui

import (
	"fmt"
	"html/template"
	"time"

	"github.com/racoon-devel/venera/internal/utils"

	"github.com/racoon-devel/venera/internal/types"
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

	case types.StatusDone:
		return "Done"

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

func relation(rel int) string {
	switch rel {
	case types.Negative:
		return "negative"
	case types.Neutral:
		return "neutral"
	case types.Positive:
		return "positive"
	default:
		return "not defined"
	}
}

func body(body int) string {
	switch body {
	case types.Fat:
		return "fat"
	case types.Sport:
		return "sport"
	case types.Thin:
		return "thin"
	default:
		return "not defined"
	}
}
