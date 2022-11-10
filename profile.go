package twitterscraper

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Global cache for user IDs
var cacheIDs sync.Map

// Profile of twitter user.
type Profile struct {
	Avatar           string
	IsNFTAvatar      bool
	Banner           string
	Biography        string
	FollowersCount   int
	FollowingCount   int
	IsPrivate        bool
	IsVerified       bool
	IsVerifiedBlue   bool
	Joined           *time.Time
	LikesCount       int
	ListedCount      int
	Location         string
	Name             string
	PinnedTweetIDs   []string
	TweetsCount      int
	URL              string
	UserID           string
	Username         string
	Website          string
	ProfessionalType string
	ProfessionalDesc string
	AffiliatesType   string
	AffiliatesDesc   string
}

type user struct {
	Data struct {
		User struct {
			Result struct {
				TypeName string `json:"__typename"`
				RestId   string `json:"rest_id"`
				// things we care about
				HasNFTAvatar   bool         `json:"has_nft_avatar"`
				IsBlueVerified bool         `json:"is_blue_verified"`
				Legacy         legacyUser   `json:"legacy"`
				Professional   professional `json:"professional"`
				Reason         string       `json:"reason"`

				Affiliates Affiliates `json:"affiliates_highlighted_label"`

				// Unavailible
				UnavailableMessage struct {
					Rtl  bool   `json:"rtl"`
					Text string `json:"text"`
				} `json:"unavailable_message"`
			} `json:"result"`
		} `json:"user"`
	} `json:"data"`
}

// GetProfile return parsed user profile.
func (s *Scraper) GetProfile(username string) (Profile, error) {
	var jsn user
	req, err := http.NewRequest("GET", "https://twitter.com/i/api/graphql/ptQPCD7NrFS_TW71Lq07nw/UserByScreenName?variables%3D%7B%22screen_name%22%3A%22"+username+"%22%2C%22withSafetyModeUserFields%22%3Atrue%2C%22withSuperFollowsUserFields%22%3Atrue%7D%26features%3D%7B%22responsive_web_twitter_blue_verified_badge_is_enabled%22%3Atrue%2C%22verified_phone_label_enabled%22%3Afalse%2C%22responsive_web_graphql_timeline_navigation_enabled%22%3Atrue%7D", nil)
	if err != nil {
		return Profile{}, err
	}

	err = s.RequestAPI(req, &jsn)
	if err != nil {
		return Profile{}, err
	}

	if jsn.Data.User.Result.Reason != "" {
		return Profile{}, fmt.Errorf("%s", jsn.Data.User.Result.Reason)
	}

	if jsn.Data.User.Result.RestId == "" {
		return Profile{}, fmt.Errorf("rest_id not found")
	}

	jsn.Data.User.Result.Legacy.IDStr = jsn.Data.User.Result.RestId

	if jsn.Data.User.Result.Legacy.ScreenName == "" {
		return Profile{}, fmt.Errorf("either @%s does not exist or is private", username)
	}

	return parseProfile(jsn), nil
}

// Deprecated: GetProfile wrapper for default scraper
func GetProfile(username string) (Profile, error) {
	return defaultScraper.GetProfile(username)
}

// GetUserIDByScreenName from API
func (s *Scraper) GetUserIDByScreenName(screenName string) (string, error) {
	id, ok := cacheIDs.Load(screenName)
	if ok {
		return id.(string), nil
	}

	profile, err := s.GetProfile(screenName)
	if err != nil {
		return "", err
	}

	cacheIDs.Store(screenName, profile.UserID)

	return profile.UserID, nil
}
