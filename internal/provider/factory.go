package provider

import (
	"fmt"
	"racoondev.tk/gitea/racoon/venera/internal/provider/export"

	"github.com/ccding/go-logging/logging"

	"racoondev.tk/gitea/racoon/venera/internal/provider/mamba"
	"racoondev.tk/gitea/racoon/venera/internal/provider/tinder"
	"racoondev.tk/gitea/racoon/venera/internal/types"
)

var (
	providers = map[string]types.Provider{
		"tinder": new(tinder.TinderProvider),
		"mamba":  new(mamba.MambaProvider),
		"export": new(export.ExportProvider),
	}
)

func SetLogger(log *logging.Logger) {
	for _, provider := range providers {
		provider.SetLogger(log)
	}
}

// GetAvailable - show all available providers
func GetAvailable() []string {
	result := make([]string, 0, len(providers))
	for id := range providers {
		result = append(result, id)
	}

	return result
}

// Get - get provider by id
func Get(id string) (types.Provider, error) {
	provider := providers[id]

	if provider == nil {
		return nil, fmt.Errorf("Provider %s not registered", id)
	}

	return provider, nil
}

// All - get all providers
func All() map[string]types.Provider {
	return providers
}
