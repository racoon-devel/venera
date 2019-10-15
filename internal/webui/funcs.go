package webui

import (
	"fmt"
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

func mod2(n int) int {
	return n % 2
}
