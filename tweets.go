package twitterscraper

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"
)

// GetTweets returns channel with tweets for a given user.
func (s *Scraper) GetTweets(ctx context.Context, user string, maxTweetsNbr int) <-chan *TweetResult {
	return getTweetTimeline(ctx, user, maxTweetsNbr, s.FetchTweets)
}

// Deprecated: GetTweets wrapper for default Scraper
func GetTweets(ctx context.Context, user string, maxTweetsNbr int) <-chan *TweetResult {
	return defaultScraper.GetTweets(ctx, user, maxTweetsNbr)
}

// FetchTweets gets tweets for a given user, via the Twitter frontend API.
func (s *Scraper) FetchTweets(user string, maxTweetsNbr int, cursor string) ([]*Tweet, string, error) {
	if maxTweetsNbr > 200 {
		maxTweetsNbr = 200
	}

	userID, err := s.GetUserIDByScreenName(user)
	if err != nil {
		return nil, "", err
	}

	req, err := s.newRequest("GET", "https://api.twitter.com/2/timeline/profile/"+userID+".json")
	if err != nil {
		return nil, "", err
	}

	q := req.URL.Query()
	q.Add("count", strconv.Itoa(maxTweetsNbr))
	q.Add("userId", userID)
	if cursor != "" {
		q.Add("cursor", cursor)
	}
	req.URL.RawQuery = q.Encode()

	var timeline timeline
	err = s.RequestAPI(req, &timeline)
	if err != nil {
		return nil, "", err
	}

	tweets, nextCursor := timeline.parseTweets()
	return tweets, nextCursor, nil
}

// GetTweet get a single tweet by ID.
func (s *Scraper) GetTweet(id string) (*Tweet, error) {
	req, err := s.newRequest("GET", "https://twitter.com/i/api/2/timeline/conversation/"+id+".json")
	if err != nil {
		return nil, err
	}

	var timeline timeline
	err = s.RequestAPI(req, &timeline)
	if err != nil {
		return nil, err
	}

	tweets, _ := timeline.parseTweets()
	for _, tweet := range tweets {
		if tweet.ID == id {
			return tweet, nil
		}
	}
	return nil, fmt.Errorf("tweet with ID %s not found", id)
}

// Deprecated: GetTweet wrapper for default Scraper
func GetTweet(id string) (*Tweet, error) {
	return defaultScraper.GetTweet(id)
}

type timelinerecursive struct {
	Errors []Err `json:"error"`
	Data   struct {
		ThreadedConvo struct {
			Instructions []instrutions `json:"instructions"`
		} `json:"threaded_conversation_with_injections_v2"`
	} `json:"data"`
}

type Err struct {
	Message string `json:"message"`
}

type instrutions struct {
	Type      string                   `json:"type"`
	Direction string                   `json:"direction"`
	Entires   []recursivetimelineentry `json:"entries"`
}

type recursivetimelineentry struct {
	EntryId   string `json:"entryId"`
	SortIndex string `json:"sortIndex"`
	content   struct {
		EntryType   string `json:"entryType"`
		TypeName    string `json:"__typename"`
		itemcontent struct {
			ItemType      string `json:"itemType"`
			TypeName      string `json:"__typename"`
			tweet_results struct {
				result struct {
					TypeName string `json:"__typename"`
					RestId   string `json:"rest_id"`
					core     struct {
						user_results struct {
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
						} `json:"user_results"`
					} `json:"core"`
				} `json:"result"`
			} `json:"tweet_results"`

			Legacy struct {
				ConversationIDStr string `json:"conversation_id_str"`
				CreatedAt         string `json:"created_at"`
				FavoriteCount     int    `json:"favorite_count"`
				FullText          string `json:"full_text"`
				Entities          struct {
					Hashtags []struct {
						Text string `json:"text"`
					} `json:"hashtags"`
					Media []struct {
						MediaURLHttps string `json:"media_url_https"`
						Type          string `json:"type"`
						URL           string `json:"url"`
					} `json:"media"`
					URLs []struct {
						ExpandedURL string `json:"expanded_url"`
						URL         string `json:"url"`
					} `json:"urls"`
				} `json:"entities"`
				ExtendedEntities struct {
					Media []struct {
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
					} `json:"media"`
				} `json:"extended_entities"`
				InReplyToStatusIDStr string    `json:"in_reply_to_status_id_str"`
				Place                Place     `json:"place"`
				ReplyCount           int       `json:"reply_count"`
				RetweetCount         int       `json:"retweet_count"`
				RetweetedStatusIDStr string    `json:"retweeted_status_id_str"`
				QuotedStatusIDStr    string    `json:"quoted_status_id_str"`
				Time                 time.Time `json:"time"`
				UserIDStr            string    `json:"user_id_str"`
			} `json:"legacy"`

			HasModeratedReplies bool `json:"hasModeratedReplies"`

			// cursor
			Value  string `json:"value"`
			Cursor string `json:"cursorType"`
		} `json:"itemContent"`
	} `json:"content"`
}

type TweetThreadTree struct {
	Store map[string]Tweet
	Root  RootTweet
}

type RootTweet struct {
	Content string
	Replies []*NodeTweet
}

type NodeTweet struct {
	Id      string
	Replies []*NodeTweet
}

func (s *Scraper) GetTweetAndRepliesRecursive(id string) ([]Tweet, error) {
	tweets := []Tweet{}

	req, err := s.newRequest("GET", "https://twitter.com/i/api/graphql/BoHLKeBvibdYDiJON1oqTg/TweetDetail?variables=%7B%22focalTweetId%22%3A%22"+id+"%22%2C%22with_rux_injections%22%3Afalse%2C%22includePromotedContent%22%3Afalse%2C%22withCommunity%22%3Afalse%2C%22withQuickPromoteEligibilityTweetFields%22%3Afalse%2C%22withBirdwatchNotes%22%3Afalse%2C%22withSuperFollowsUserFields%22%3Afalse%2C%22withDownvotePerspective%22%3Afalse%2C%22withReactionsMetadata%22%3Afalse%2C%22withReactionsPerspective%22%3Afalse%2C%22withSuperFollowsTweetFields%22%3Afalse%2C%22withVoice%22%3Afalse%2C%22withV2Timeline%22%3Atrue%7D%26features%3D%7B%22responsive_web_twitter_blue_verified_badge_is_enabled%22%3Atrue%2C%22verified_phone_label_enabled%22%3Afalse%2C%22responsive_web_graphql_timeline_navigation_enabled%22%3Atrue%2C%22unified_cards_ad_metadata_container_dynamic_card_content_query_enabled%22%3Atrue%2C%22tweetypie_unmention_optimization_enabled%22%3Atrue%2C%22responsive_web_uc_gql_enabled%22%3Atrue%2C%22vibe_api_enabled%22%3Atrue%2C%22responsive_web_edit_tweet_api_enabled%22%3Atrue%2C%22graphql_is_translatable_rweb_tweet_is_translatable_enabled%22%3Atrue%2C%22standardized_nudges_misinfo%22%3Atrue%2C%22tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled%22%3Afalse%2C%22interactive_text_enabled%22%3Atrue%2C%22responsive_web_text_conversations_enabled%22%3Afalse%2C%22responsive_web_enhance_cards_enabled%22%3Atrue%7D")
	if err != nil {
		return tweets, err
	}

	timelines := []timelinerecursive{}

	var firstjsn timelinerecursive
	err = s.RequestAPI(req, &firstjsn)
	if err != nil {
		return tweets, err
	}

	if len(firstjsn.Errors) > 0 {
		return tweets, fmt.Errorf("%s", firstjsn.Errors[0].Message)
	}

	timelines = append(timelines, firstjsn)

	// find cursor recursively
	for {
		last_timeline := timelines[len(timelines)-1]

		if len(last_timeline.Data.ThreadedConvo.Instructions) <= 0 {
			break
		}

		if len(last_timeline.Data.ThreadedConvo.Instructions[0].Entires) <= 0 {
			break
		}

		if last_timeline.Data.ThreadedConvo.Instructions[0].Entires[len(last_timeline.Data.ThreadedConvo.Instructions[0].Entires)-1].content.itemcontent.Cursor != "" {
			// requsts with cursor
			cursor := url.QueryEscape(last_timeline.Data.ThreadedConvo.Instructions[0].Entires[len(last_timeline.Data.ThreadedConvo.Instructions[0].Entires)-1].content.itemcontent.Cursor)
			req, err = s.newRequest("GET", "https://twitter.com/i/api/graphql/BoHLKeBvibdYDiJON1oqTg/TweetDetail?variables%3D%7B%22focalTweetId%22%3A%22"+id+"%22%2C%22cursor%22%3A%22"+cursor+"%22%2C%22referrer%22%3A%22messages%22%2C%22with_rux_injections%22%3Afalse%2C%22includePromotedContent%22%3Afa%3Bse%2C%22withCommunity%22%3Afalse%2C%22withQuickPromoteEligibilityTweetFields%22%3Afalse%2C%22withBirdwatchNotes%22%3Afalse%2C%22withSuperFollowsUserFields%22%3Afalse%2C%22withDownvotePerspective%22%3Afalse%2C%22withReactionsMetadata%22%3Afalse%2C%22withReactionsPerspective%22%3Afalse%2C%22withSuperFollowsTweetFields%22%3Afalse%2C%22withVoice%22%3Atrue%2C%22withV2Timeline%22%3Atrue%7D%26features%3D%7B%22responsive_web_twitter_blue_verified_badge_is_enabled%22%3Atrue%2C%22verified_phone_label_enabled%22%3Afalse%2C%22responsive_web_graphql_timeline_navigation_enabled%22%3Atrue%2C%22unified_cards_ad_metadata_container_dynamic_card_content_query_enabled%22%3Atrue%2C%22tweetypie_unmention_optimization_enabled%22%3Atrue%2C%22responsive_web_uc_gql_enabled%22%3Atrue%2C%22vibe_api_enabled%22%3Atrue%2C%22responsive_web_edit_tweet_api_enabled%22%3Atrue%2C%22graphql_is_translatable_rweb_tweet_is_translatable_enabled%22%3Atrue%2C%22standardized_nudges_misinfo%22%3Atrue%2C%22tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled%22%3Afalse%2C%22interactive_text_enabled%22%3Atrue%2C%22responsive_web_text_conversations_enabled%22%3Afalse%2C%22responsive_web_enhance_cards_enabled%22%3Atrue%7D")
			if err != nil {
				return tweets, err
			}

			var jsn timelinerecursive
			err = s.RequestAPI(req, &jsn)
			if err != nil {
				return tweets, err
			}

			if len(jsn.Errors) > 0 {
				return tweets, fmt.Errorf("%s", jsn.Errors[0].Message)
			}

			timelines = append(timelines, jsn)

		} else {
			break
		}
	}

	for timeline := timelines {
		// get tweets

		// parse entries
		
	}

}
