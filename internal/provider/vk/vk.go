package vk

import "racoondev.tk/gitea/racoon/venera/internal/types"

type searchSettings struct {
	User     string
	Password string
	AgeFrom  uint
	AgeTo    uint
	Keywords []string
	Likes    []string
	Dislikes []string
}

// VkProvider - provider for searching people in VK
type VkProvider struct {
}

func (ctx VkProvider) RestoreSearchSession(state string) types.SearchSession {
	var session vkSearchSession
	session.LoadState(state)
	return &session
}
