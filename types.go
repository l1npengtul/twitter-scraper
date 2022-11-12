package twitterscraper

import "time"

type (
	// Video type.
	Video struct {
		ID      string
		Preview string
		URL     string
	}

	// Tweet type.
	Tweet struct {
		ID               string
		EditIds          []string
		Hashtags         []string
		ReplyingTo       string
		QuoteRetweetId   string
		IsReply          bool
		IsRetweet        bool
		PermanentURL     string
		Media            []Media
		Card             Card
		Mentions         map[string]string
		Text             string
		TimeParsed       time.Time
		Timestamp        int64
		UserID           string
		Username         string
		SensitiveContent bool
		Likes            int
		Retweets         int
		QuoteRetweets    int
		Replies          int
		IsTombstone      bool
	}

	// ProfileResult of scrapping.
	ProfileResult struct {
		Profile
		Error error
	}

	// TweetResult of scrapping.
	TweetResult struct {
		Tweet
		Error error
	}

	legacyUser struct {
		CreatedAt   string `json:"created_at"`
		Description string `json:"description"`
		Entities    struct {
			URL struct {
				Urls []struct {
					ExpandedURL string `json:"expanded_url"`
				} `json:"urls"`
			} `json:"url"`
		} `json:"entities"`
		FavouritesCount      int      `json:"favourites_count"`
		FollowersCount       int      `json:"followers_count"`
		FriendsCount         int      `json:"friends_count"`
		IDStr                string   `json:"id_str"`
		ListedCount          int      `json:"listed_count"`
		Name                 string   `json:"name"`
		Location             string   `json:"location"`
		PinnedTweetIdsStr    []string `json:"pinned_tweet_ids_str"`
		ProfileBannerURL     string   `json:"profile_banner_url"`
		ProfileImageURLHTTPS string   `json:"profile_image_url_https"`
		Protected            bool     `json:"protected"`
		ScreenName           string   `json:"screen_name"`
		StatusesCount        int      `json:"statuses_count"`
		Verified             bool     `json:"verified"`
	}

	professional struct {
		ProfessionalType string     `json:"professional_type"`
		RestId           string     `json:"rest_id"`
		Category         []category `json:"category"`
	}

	category struct {
		Id       int    `json:"id"`
		Name     string `json:"name"`
		IconName string `json:"icon_name"`
	}

	Affiliates struct {
		Label struct {
			Url struct {
				Url     string `json:"url"`
				UrlType string `json:"urlType"`
			} `json:"url"`
			Badge struct {
				Url string `json:"url"`
			} `json:"badge"`
			Description string `json:"description"`
		} `json:"label"`
	}

	Place struct {
		ID          string `json:"id"`
		PlaceType   string `json:"place_type"`
		Name        string `json:"name"`
		FullName    string `json:"full_name"`
		CountryCode string `json:"country_code"`
		Country     string `json:"country"`
		BoundingBox struct {
			Type        string        `json:"type"`
			Coordinates [][][]float64 `json:"coordinates"`
		} `json:"bounding_box"`
	}

	Media struct {
		IDStr                    string `json:"id_str"`
		MediaURLHttps            string `json:"media_url_https"`
		ExtSensitiveMediaWarning struct {
			AdultContent    bool `json:"adult_content"`
			GraphicViolence bool `json:"graphic_violence"`
			Other           bool `json:"other"`
		} `json:"ext_sensitive_media_warning"`
		Type      string `json:"type"`
		URL       string `json:"url"`
		VideoInfo struct {
			Variants []struct {
				Bitrate int    `json:"bitrate,omitempty"`
				URL     string `json:"url"`
			} `json:"variants"`
		} `json:"video_info"`
	}

	fetchProfileFunc func(query string, maxProfilesNbr int, cursor string) ([]*Profile, string, error)
	fetchTweetFunc   func(query string, maxTweetsNbr int, cursor string) ([]*Tweet, string, error)
)
