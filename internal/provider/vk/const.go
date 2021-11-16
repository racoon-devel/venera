package vk

import "time"

const (
	appId = 6122372

	sexWoman           = 1
	statusNotMarried   = 1
	statusActiveSearch = 6

	maxPhotosLimit          = 10
	vkSearchResultsLimit    = 1000
	searchRequestBatch      = vkSearchResultsLimit
	maxGroupSearchLimit     = 50
	expiredAccountThreshold = 15 * 24 * time.Hour

	searchFields          = "sex,personal,interests,music,movies,tv,games,about,books,quotes,occupation,photo_max_orig,last_seen,bdate"
	friendsLimitThreshold = 500
)
