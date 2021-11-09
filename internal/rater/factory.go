package rater

import (
	"github.com/ccding/go-logging/logging"
	"github.com/racoon-devel/venera/internal/types"
)

const (
	IgnorePerson = -100
)

type raterCreator func(configuration string, log *logging.Logger, settings *types.SearchSettings) types.Rater

var (
	factoryMethods = map[string]raterCreator{
		"ml":         createMlRater,
		"default+ml": createCompositeRater,
		"default":    createDefaultRater,
	}
)

func NewRater(raterID string, configuration string, log *logging.Logger, settings *types.SearchSettings) types.Rater {
	log.Debugf("Instancing rater '%s'", raterID)

	creator, ok := factoryMethods[raterID]
	if !ok {
		log.Warnf("Unknown rater: %s", raterID)
		return createDefaultRater(configuration, log, settings)
	}

	return creator(configuration, log, settings)
}

func GetRaters() []string {
	all := make([]string, 0)
	for k, _ := range factoryMethods {
		all = append(all, k)
	}

	return all
}

func createMlRater(configuration string, log *logging.Logger, settings *types.SearchSettings) types.Rater {
	rater := &mlRater{}
	rater.Init(log, settings)
	return rater
}

func createCompositeRater(configuration string, log *logging.Logger, settings *types.SearchSettings) types.Rater {
	def := &defaultRater{configName: configuration}
	def.Init(log, settings)

	ml := &mlRater{}
	ml.Init(log, settings)

	def.Next(ml)

	return def
}

func createDefaultRater(configuration string, log *logging.Logger, settings *types.SearchSettings) types.Rater {
	rater := &defaultRater{configName: configuration}
	rater.Init(log, settings)
	return rater
}
