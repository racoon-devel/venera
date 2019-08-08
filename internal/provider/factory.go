package provider

import (
	"fmt"
	"net/http"

	"racoondev.tk/gitea/racoon/venera/internal/provider/vk"
)

// Provider - object for searching people in some social network
type Provider interface {
	NewTaskPageHandler(w http.ResponseWriter, r *http.Request)
}

var (
	providers = map[string]Provider{
		"vk": new(vk.VkProvider),
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
func Get(id string) (Provider, error) {
	provider := providers[id]

	if provider == nil {
		return nil, fmt.Errorf("Provider %s not registered", id)
	}

	return provider, nil
}

// All - get all providers
func All() map[string]Provider {
	return providers
}
