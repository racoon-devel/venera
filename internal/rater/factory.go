package rater

import (
	"github.com/ccding/go-logging/logging"
	"racoondev.tk/gitea/racoon/venera/internal/types"
)

func NewRater(raterID string, log *logging.Logger, settings *types.SearchSettings) types.Rater {
	switch raterID {
	case "default":
		rater := &defaultRater{}
		rater.Init(log, settings)
		return rater

	default:
		log.Errorf("Rater '%s' doesn't exists", raterID)
		return nil
	}
}
