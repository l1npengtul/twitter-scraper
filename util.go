package twitterscraper

import (
	"context"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

var (
	reHashtag    = regexp.MustCompile(`\B(\#\S+\b)`)
	reTwitterURL = regexp.MustCompile(`https:(\/\/t\.co\/([A-Za-z0-9]|[A-Za-z]){10})`)
	reUsername   = regexp.MustCompile(`\B(\@\S{1,15}\b)`)
)

func (s *Scraper) newRequest(method string, url string) (*http.Request, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("include_profile_interstitial_type", "1")
	q.Add("include_blocking", "1")
	q.Add("include_blocked_by", "1")
	q.Add("include_followed_by", "1")
	q.Add("include_want_retweets", "1")
	q.Add("include_mute_edge", "1")
	q.Add("include_can_dm", "1")
	q.Add("include_can_media_tag", "1")
	q.Add("include_ext_has_nft_avatar", "1")
	q.Add("skip_status", "1")
	q.Add("cards_platform", "Web-12")
	q.Add("include_cards", "1")
	q.Add("include_ext_alt_text", "true")
	q.Add("include_quote_count", "true")
	q.Add("include_reply_count", "1")
	q.Add("tweet_mode", "extended")
	q.Add("include_entities", "true")
	q.Add("include_user_entities", "true")
	q.Add("include_ext_media_color", "true")
	q.Add("include_ext_media_availability", "true")
	q.Add("include_ext_sensitive_media_warning", "true")
	q.Add("send_error_codes", "true")
	q.Add("simple_quoted_tweet", "true")
	q.Add("include_tweet_replies", strconv.FormatBool(s.includeReplies))
	q.Add("ext", "mediaStats,highlightedLabel,hasNftAvatar,voiceInfo,superFollowMetadata")
	req.URL.RawQuery = q.Encode()

	return req, nil
}

func getUserTimeline(ctx context.Context, query string, maxProfilesNbr int, fetchFunc fetchProfileFunc) <-chan *ProfileResult {
	channel := make(chan *ProfileResult)
	go func(query string) {
		defer close(channel)
		var nextCursor string
		profilesNbr := 0
		for profilesNbr < maxProfilesNbr {
			select {
			case <-ctx.Done():
				channel <- &ProfileResult{Error: ctx.Err()}
				return
			default:
			}

			profiles, next, err := fetchFunc(query, maxProfilesNbr, nextCursor)
			if err != nil {
				channel <- &ProfileResult{Error: err}
				return
			}

			if len(profiles) == 0 {
				break
			}

			for _, profile := range profiles {
				select {
				case <-ctx.Done():
					channel <- &ProfileResult{Error: ctx.Err()}
					return
				default:
				}

				if profilesNbr < maxProfilesNbr {
					nextCursor = next
					channel <- &ProfileResult{Profile: *profile}
				} else {
					break
				}
				profilesNbr++
			}
		}
	}(query)
	return channel
}

func getTweetTimeline(ctx context.Context, query string, maxTweetsNbr int, fetchFunc fetchTweetFunc) <-chan *TweetResult {
	channel := make(chan *TweetResult)
	go func(query string) {
		defer close(channel)
		var nextCursor string
		tweetsNbr := 0
		for tweetsNbr < maxTweetsNbr {
			select {
			case <-ctx.Done():
				channel <- &TweetResult{Error: ctx.Err()}
				return
			default:
			}

			tweets, next, err := fetchFunc(query, maxTweetsNbr, nextCursor)
			if err != nil {
				channel <- &TweetResult{Error: err}
				return
			}

			if len(tweets) == 0 {
				break
			}

			for _, tweet := range tweets {
				select {
				case <-ctx.Done():
					channel <- &TweetResult{Error: ctx.Err()}
					return
				default:
				}

				if tweetsNbr < maxTweetsNbr {
					if tweet.IsPin && nextCursor != "" {
						continue
					}
					nextCursor = next
					channel <- &TweetResult{Tweet: *tweet}
				} else {
					break
				}
				tweetsNbr++
			}
		}
	}(query)
	return channel
}

func parseProfile(user user) Profile {
	profile := Profile{
		Avatar:           user.Data.User.Result.Legacy.ProfileImageURLHTTPS,
		IsNFTAvatar:      user.Data.User.Result.HasNFTAvatar,
		Banner:           user.Data.User.Result.Legacy.ProfileBannerURL,
		Biography:        user.Data.User.Result.Legacy.Description,
		FollowersCount:   user.Data.User.Result.Legacy.FollowersCount,
		FollowingCount:   user.Data.User.Result.Legacy.FriendsCount,
		IsPrivate:        user.Data.User.Result.Legacy.Protected,
		IsVerified:       user.Data.User.Result.Legacy.Verified,
		IsVerifiedBlue:   user.Data.User.Result.IsBlueVerified,
		LikesCount:       user.Data.User.Result.Legacy.FavouritesCount,
		ListedCount:      user.Data.User.Result.Legacy.ListedCount,
		Location:         user.Data.User.Result.Legacy.Location,
		Name:             user.Data.User.Result.Legacy.Name,
		PinnedTweetIDs:   user.Data.User.Result.Legacy.PinnedTweetIdsStr,
		TweetsCount:      user.Data.User.Result.Legacy.StatusesCount,
		URL:              "https://twitter.com/" + user.Data.User.Result.Legacy.ScreenName,
		UserID:           user.Data.User.Result.Legacy.IDStr,
		Username:         user.Data.User.Result.Legacy.ScreenName,
		ProfessionalType: user.Data.User.Result.Professional.ProfessionalType,
		AffiliatesType:   user.Data.User.Result.Affiliates.Label.Url.Url,
		AffiliatesDesc:   user.Data.User.Result.Affiliates.Label.Description,
	}

	tm, err := time.Parse(time.RubyDate, user.Data.User.Result.Legacy.CreatedAt)
	if err == nil {
		tm = tm.UTC()
		profile.Joined = &tm
	}

	if len(user.Data.User.Result.Legacy.Entities.URL.Urls) > 0 {
		profile.Website = user.Data.User.Result.Legacy.Entities.URL.Urls[0].ExpandedURL
	}

	if len(user.Data.User.Result.Professional.Category) > 0 {
		profile.ProfessionalDesc = user.Data.User.Result.Professional.Category[0].Name
	}

	return profile
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
