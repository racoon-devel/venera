package provider

import (
	"fmt"

	"racoondev.tk/gitea/racoon/venera/internal/provider/tinder"
	"racoondev.tk/gitea/racoon/venera/internal/provider/vk"
	"racoondev.tk/gitea/racoon/venera/internal/types"
)

var (
	providers = map[string]types.Provider{
		"vk":     new(vk.VkProvider),
		"tinder": new(tinder.TinderProvider),
	}
)

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
