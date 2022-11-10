package twitterscraper

import (
	"context"
	"fmt"
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
	ThreadedConvo struct {
		Instructions struct {
			Entries struct {
				Type string
			}
			Terminate struct {
				Type      string
				Direction string
			}
		}
	}
}

type recursivetimelineentry struct {
	EntryId   string
	SortIndex string
	content   struct {
		EntryType string
		TypeName  string
	}
}

type items struct {
	EntryId string
	item    struct {
		itemcontent struct {
			ItemType      string
			TypeName      string
			tweet_results struct {
				result struct {
					TypeName string
					RestId   string
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
						}
					}
				}
			}

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
		}
	}
}

func GetTweetAndRepliesRecursive(id string) ([]Tweet, error) {
	tweets := []Tweet{}

}
