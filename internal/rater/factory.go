package rater

import (
	"github.com/ccding/go-logging/logging"
	"racoondev.tk/gitea/racoon/venera/internal/types"
)

const IgnorePerson = -100

func NewRater(raterID string, configuration string, log *logging.Logger, settings *types.SearchSettings) types.Rater {
	switch raterID {
	case "default":
		rater := &defaultRater{configName: configuration}
		rater.Init(log, settings)
		return rater

	default:
		log.Errorf("Rater '%s' doesn't exists", raterID)
		return nil
	}
}
