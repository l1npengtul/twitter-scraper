package twitterscraper

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type timelinerecursive struct {
	Errors []Err `json:"error"`
	Data   struct {
		ThreadedConvo struct {
			Instructions []instrutions `json:"instructions"`
		} `json:"threaded_conversation_with_injections_v2"`
	} `json:"data"`
}

type Err struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type instrutions struct {
	Type      string                   `json:"type"`
	Direction string                   `json:"direction"`
	Entries   []recursivetimelineentry `json:"entries"`
}

type recursivetimelineentry struct {
	EntryId   string `json:"entryId"`
	SortIndex string `json:"sortIndex"`
	Content   struct {
		EntryType   string      `json:"entryType"`
		TypeName    string      `json:"__typename"`
		ItemContent itemcontent `json:"itemContent"`
		Items       []items     `json:"items"`
	} `json:"content"`
}

type items struct {
	EntryId string `json:"entryId"`
	Item    struct {
		ItemContent itemcontent `json:"itemContent"`
	} `json:"item"`
}

type itemcontent struct {
	ItemType     string `json:"itemType"`
	TypeName     string `json:"__typename"`
	TweetResults struct {
		Result struct {
			TypeName string `json:"__typename"`
			RestId   string `json:"rest_id"`
			Core     struct {
				UserResults struct {
					Result UserResult `json:"result"`
				} `json:"user_results"`
			} `json:"core"`

			Card Card `json:"card"`

			EditControl struct {
				InitialTweetId string   `json:"initial_tweet_id"`
				EditIds        []string `json:"edit_tweet_ids"`
				EditableUntil  string   `json:"editable_until_msecs"`
				IsEditable     bool     `json:"is_edit_eligible"`
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
					UserMentions []struct {
						IdStr      string `json:"id_str"`
						Name       string `json:"name"`
						ScreenName string `json:"screen_name"`
					} `json:"user_mentions"`
				} `json:"entities"`
				ExtendedEntities struct {
					Media []Media `json:"media"`
				} `json:"extended_entities"`
				InReplyToStatusIDStr string    `json:"in_reply_to_status_id_str"`
				Place                Place     `json:"place"`
				ReplyCount           int       `json:"reply_count"`
				RetweetCount         int       `json:"retweet_count"`
				RetweetedStatusIDStr string    `json:"retweeted_status_id_str"`
				QuoteCount           int       `json:"quote_count"`
				PossiblySensitive    bool      `json:"possibly_sensitive"`
				QuotedStatusIDStr    string    `json:"quoted_status_id_str"`
				Time                 time.Time `json:"time"`
				UserIDStr            string    `json:"user_id_str"`
			} `json:"legacy"`
			HasModeratedReplies bool `json:"hasModeratedReplies"`
		} `json:"result"`
	} `json:"tweet_results"`

	// cursor
	Value  string `json:"value"`
	Cursor string `json:"cursorType"`
}

type Card struct {
	RestId string
	Legacy struct {
		BindingValues []struct {
			Key   string `json:"key"`
			Value struct {
				StringValue string `json:"string_value"`
				Type        string `json:"type"`
			} `json:"value"`
		} `json:"binding_values"`
	} `json:"legacy"`
}

type TweetThreadTree struct {
	Store map[string]Tweet
	Root  NodeTweet
}

type NodeTweet struct {
	Id      string
	Replies []NodeTweet
}

func (self *TweetThreadTree) GetOrErr(id string) (Tweet, error) {
	val, exists := self.Store[id]
	if exists {
		return val, nil
	}
	return Tweet{}, fmt.Errorf("doesnt Exist")
}

func (self *TweetThreadTree) RepliesOfTweet(id string) int {
	if id == self.Root.Id {
		return len(self.Root.Replies)
	}

	var nodes []NodeTweet
	nodes = append(nodes, self.Root)

	for {
		if len(nodes) == 0 {
			break
		}

		topnode := nodes[0]
		//x, a = a[len(a)-1], a[:len(a)-1]	}
		_, nodes = nodes[len(nodes)-1], nodes[:len(nodes)-1]
		if topnode.Id == id {
			return len(topnode.Replies)
		}
		nodes = append(nodes, topnode.Replies...)
	}
	return -1
}

func (self *TweetThreadTree) InsertRootNode(tweet Tweet) {
	id := tweet.ID
	self.Store[id] = tweet

	self.Root.Id = id
}

func (self *TweetThreadTree) InsertTweet(tweet Tweet, replying_to string) bool {
	id := tweet.ID
	self.Store[id] = tweet

	if replying_to == self.Root.Id {
		self.Root.Replies = append(self.Root.Replies, NodeTweet{Id: tweet.ID})
		return true
	}

	var nodes []NodeTweet
	nodes = append(nodes, self.Root)

	for {
		if len(nodes) == 0 {
			break
		}

		topnode := nodes[0]
		//x, a = a[len(a)-1], a[:len(a)-1]	}
		_, nodes = nodes[len(nodes)-1], nodes[:len(nodes)-1]
		if topnode.Id == replying_to {
			topnode.Replies = append(topnode.Replies, NodeTweet{Id: tweet.ID})
			return true
		}
		nodes = append(nodes, topnode.Replies...)
	}
	return false
}

func (self *TweetThreadTree) Flatten() []Tweet {
	tweets := make([]Tweet, 0, len(self.Store))
	for _, tweet := range self.Store {
		tweets = append(tweets, tweet)
	}
	return tweets

}

type contents struct {
	ItemContent itemcontent
	Entry       string
}

func (s *Scraper) GetTweetAndRepliesRecursive(id string) ([]Tweet, map[string]Profile, error) {
	var tweets []Tweet
	var users map[string]Profile
	var tlContents []contents

	req, err := s.newRequest("GET", "https://twitter.com/i/api/graphql/BoHLKeBvibdYDiJON1oqTg/TweetDetail?variables=%7B%22focalTweetId%22%3A%22"+id+"%22%2C%22with_rux_injections%22%3Afalse%2C%22includePromotedContent%22%3Afalse%2C%22withCommunity%22%3Afalse%2C%22withQuickPromoteEligibilityTweetFields%22%3Afalse%2C%22withBirdwatchNotes%22%3Afalse%2C%22withSuperFollowsUserFields%22%3Afalse%2C%22withDownvotePerspective%22%3Afalse%2C%22withReactionsMetadata%22%3Afalse%2C%22withReactionsPerspective%22%3Afalse%2C%22withSuperFollowsTweetFields%22%3Afalse%2C%22withVoice%22%3Afalse%2C%22withV2Timeline%22%3Atrue%7D%26features%3D%7B%22responsive_web_twitter_blue_verified_badge_is_enabled%22%3Atrue%2C%22verified_phone_label_enabled%22%3Afalse%2C%22responsive_web_graphql_timeline_navigation_enabled%22%3Atrue%2C%22unified_cards_ad_metadata_container_dynamic_card_content_query_enabled%22%3Atrue%2C%22tweetypie_unmention_optimization_enabled%22%3Atrue%2C%22responsive_web_uc_gql_enabled%22%3Atrue%2C%22vibe_api_enabled%22%3Atrue%2C%22responsive_web_edit_tweet_api_enabled%22%3Atrue%2C%22graphql_is_translatable_rweb_tweet_is_translatable_enabled%22%3Atrue%2C%22standardized_nudges_misinfo%22%3Atrue%2C%22tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled%22%3Afalse%2C%22interactive_text_enabled%22%3Atrue%2C%22responsive_web_text_conversations_enabled%22%3Afalse%2C%22responsive_web_enhance_cards_enabled%22%3Atrue%7D")
	if err != nil {
		return tweets, users, err
	}

	var firstjsn timelinerecursive
	err = s.RequestAPI(req, &firstjsn)
	if err != nil {
		return tweets, users, err
	}

	if len(firstjsn.Errors) > 0 {
		if firstjsn.Errors[0].Code != 37 {
			return tweets, users, fmt.Errorf("%s", firstjsn.Errors[0].Message)
		}
	}

	if len(firstjsn.Data.ThreadedConvo.Instructions) < 1 {
		return tweets, users, fmt.Errorf("no Instruction")
	}

	cursorTop := ""

	for {
		for _, entry := range firstjsn.Data.ThreadedConvo.Instructions[0].Entries {
			if strings.HasPrefix(entry.EntryId, "cursor-top") {
				cursorTop = entry.Content.ItemContent.Value
				break
			}
		}

		if cursorTop != "" {
			req, err = s.newRequest("GET", "https://twitter.com/i/api/graphql/BoHLKeBvibdYDiJON1oqTg/TweetDetail?variables%3D%7B%22focalTweetId%22%3A%22"+id+"%22%2C%22cursor%22%3A%22"+cursorTop+"%22%2C%22referrer%22%3A%22messages%22%2C%22with_rux_injections%22%3Afalse%2C%22includePromotedContent%22%3Afa%3Bse%2C%22withCommunity%22%3Afalse%2C%22withQuickPromoteEligibilityTweetFields%22%3Afalse%2C%22withBirdwatchNotes%22%3Afalse%2C%22withSuperFollowsUserFields%22%3Afalse%2C%22withDownvotePerspective%22%3Afalse%2C%22withReactionsMetadata%22%3Afalse%2C%22withReactionsPerspective%22%3Afalse%2C%22withSuperFollowsTweetFields%22%3Afalse%2C%22withVoice%22%3Atrue%2C%22withV2Timeline%22%3Atrue%7D%26features%3D%7B%22responsive_web_twitter_blue_verified_badge_is_enabled%22%3Atrue%2C%22verified_phone_label_enabled%22%3Afalse%2C%22responsive_web_graphql_timeline_navigation_enabled%22%3Atrue%2C%22unified_cards_ad_metadata_container_dynamic_card_content_query_enabled%22%3Atrue%2C%22tweetypie_unmention_optimization_enabled%22%3Atrue%2C%22responsive_web_uc_gql_enabled%22%3Atrue%2C%22vibe_api_enabled%22%3Atrue%2C%22responsive_web_edit_tweet_api_enabled%22%3Atrue%2C%22graphql_is_translatable_rweb_tweet_is_translatable_enabled%22%3Atrue%2C%22standardized_nudges_misinfo%22%3Atrue%2C%22tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled%22%3Afalse%2C%22interactive_text_enabled%22%3Atrue%2C%22responsive_web_text_conversations_enabled%22%3Afalse%2C%22responsive_web_enhance_cards_enabled%22%3Atrue%7D")
			if err != nil {
				return tweets, users, err
			}

			var frontmatterjsn timelinerecursive
			err = s.RequestAPI(req, &frontmatterjsn)
			if err != nil {
				return tweets, users, err
			}

			if len(frontmatterjsn.Errors) > 0 {
				if frontmatterjsn.Errors[0].Code != 37 {
					return tweets, users, fmt.Errorf("%s", frontmatterjsn.Errors[0].Message)
				}
			}

			if len(frontmatterjsn.Data.ThreadedConvo.Instructions) < 1 {
				return tweets, users, fmt.Errorf("no instructions")
			}

			convoint := 0

			for i, inst := range frontmatterjsn.Data.ThreadedConvo.Instructions {
				if inst.Type == "TimelineAddEntries" {
					convoint = i
					break
				}
			}

			var inttwts []contents

			for _, entry := range frontmatterjsn.Data.ThreadedConvo.Instructions[convoint].Entries {
				if strings.HasPrefix(entry.EntryId, "conversationthread") {

					for _, item := range entry.Content.Items {
						inttwts = append(inttwts, contents{
							ItemContent: item.Item.ItemContent,
							Entry:       item.EntryId,
						})
					}
				} else {
					inttwts = append(inttwts, contents{
						ItemContent: entry.Content.ItemContent,
						Entry:       entry.EntryId,
					})
				}
			}

			tlContents = append(inttwts, tlContents...)
		} else {
			break
		}
	}

	var aaatwts []contents

	cvi := 0

	for i, inst := range firstjsn.Data.ThreadedConvo.Instructions {
		if inst.Type == "TimelineAddEntries" {
			cvi = i
			break
		}
	}

	for _, entry := range firstjsn.Data.ThreadedConvo.Instructions[cvi].Entries {
		if strings.HasPrefix(entry.EntryId, "conversationthread") {

			for _, item := range entry.Content.Items {
				aaatwts = append(aaatwts, contents{
					ItemContent: item.Item.ItemContent,
					Entry:       item.EntryId,
				})
			}
		} else {
			aaatwts = append(aaatwts, contents{
				ItemContent: entry.Content.ItemContent,
				Entry:       entry.EntryId,
			})
		}
	}

	tlContents = append(aaatwts, tlContents...)

	cursorBottom := ""

	for {
		for _, entry := range firstjsn.Data.ThreadedConvo.Instructions[cvi].Entries {
			if strings.HasPrefix(entry.EntryId, "cursor-bottom") {
				cursorBottom = entry.Content.ItemContent.Value
				break
			}
		}

		if cursorBottom != "" {
			req, err = s.newRequest("GET", "https://twitter.com/i/api/graphql/BoHLKeBvibdYDiJON1oqTg/TweetDetail?variables%3D%7B%22focalTweetId%22%3A%22"+id+"%22%2C%22cursor%22%3A%22"+cursorBottom+"%22%2C%22referrer%22%3A%22messages%22%2C%22with_rux_injections%22%3Afalse%2C%22includePromotedContent%22%3Afa%3Bse%2C%22withCommunity%22%3Afalse%2C%22withQuickPromoteEligibilityTweetFields%22%3Afalse%2C%22withBirdwatchNotes%22%3Afalse%2C%22withSuperFollowsUserFields%22%3Afalse%2C%22withDownvotePerspective%22%3Afalse%2C%22withReactionsMetadata%22%3Afalse%2C%22withReactionsPerspective%22%3Afalse%2C%22withSuperFollowsTweetFields%22%3Afalse%2C%22withVoice%22%3Atrue%2C%22withV2Timeline%22%3Atrue%7D%26features%3D%7B%22responsive_web_twitter_blue_verified_badge_is_enabled%22%3Atrue%2C%22verified_phone_label_enabled%22%3Afalse%2C%22responsive_web_graphql_timeline_navigation_enabled%22%3Atrue%2C%22unified_cards_ad_metadata_container_dynamic_card_content_query_enabled%22%3Atrue%2C%22tweetypie_unmention_optimization_enabled%22%3Atrue%2C%22responsive_web_uc_gql_enabled%22%3Atrue%2C%22vibe_api_enabled%22%3Atrue%2C%22responsive_web_edit_tweet_api_enabled%22%3Atrue%2C%22graphql_is_translatable_rweb_tweet_is_translatable_enabled%22%3Atrue%2C%22standardized_nudges_misinfo%22%3Atrue%2C%22tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled%22%3Afalse%2C%22interactive_text_enabled%22%3Atrue%2C%22responsive_web_text_conversations_enabled%22%3Afalse%2C%22responsive_web_enhance_cards_enabled%22%3Atrue%7D")
			if err != nil {
				return tweets, users, err
			}

			var backmatterjsn timelinerecursive
			err = s.RequestAPI(req, &backmatterjsn)
			if err != nil {
				return tweets, users, err
			}

			if len(backmatterjsn.Errors) > 0 {
				if backmatterjsn.Errors[0].Code != 37 {
					return tweets, users, fmt.Errorf("%s", backmatterjsn.Errors[0].Message)
				}
			}

			if len(backmatterjsn.Data.ThreadedConvo.Instructions) < 1 {
				return tweets, users, fmt.Errorf("no instructions")
			}

			convoint := 0

			for i, inst := range backmatterjsn.Data.ThreadedConvo.Instructions {
				if inst.Type == "TimelineAddEntries" {
					convoint = i
					break
				}
			}

			var inttwts []contents

			for _, entry := range backmatterjsn.Data.ThreadedConvo.Instructions[convoint].Entries {
				if strings.HasPrefix(entry.EntryId, "conversationthread") {

					for _, item := range entry.Content.Items {
						inttwts = append(inttwts, contents{
							ItemContent: item.Item.ItemContent,
							Entry:       item.EntryId,
						})
					}
				} else {
					inttwts = append(inttwts, contents{
						ItemContent: entry.Content.ItemContent,
						Entry:       entry.EntryId,
					})
				}
			}

			tlContents = append(inttwts, tlContents...)
		} else {
			break
		}
	}

	for _, possibleTweet := range tlContents {
		if possibleTweet.ItemContent.TweetResults.Result.TypeName == "TweetTombstone" {
			// find tweet id by parsing the entry ID
			splits := strings.Split(possibleTweet.Entry, "-")
			// get the last part as int
			if len(splits) < 4 {
				continue
			}
			tweetId := splits[4]
			tweets = append(tweets, Tweet{
				ID:          tweetId,
				TimeParsed:  time.Now(),
				IsTombstone: true,
			})
			continue
		}

		twt, err := ItemContentToTweet(possibleTweet.ItemContent)

		if err != nil {
			continue
		}
		tweets = append(tweets, twt)

		// find users
		if possibleTweet.ItemContent.TweetResults.Result.Core.UserResults.Result.RestId != "" {
			prf := parseProfile(possibleTweet.ItemContent.TweetResults.Result.Core.UserResults.Result)
			users[possibleTweet.ItemContent.TweetResults.Result.Core.UserResults.Result.RestId] = prf
		}
	}

	return tweets, users, nil
}

// FetchTweets gets tweets for a given user, via the Twitter frontend API.
func (s *Scraper) FetchTweets(user string) ([]Tweet, error) {
	var tlContents []contents
	var tweets []Tweet

	req, err := s.newRequest("GET", "https://twitter.com/i/api/graphql/BoHLKeBvibdYDiJON1oqTg/TweetDetail?variables=%7B%22focalTweetId%22%3A%22"+id+"%22%2C%22with_rux_injections%22%3Afalse%2C%22includePromotedContent%22%3Afalse%2C%22withCommunity%22%3Afalse%2C%22withQuickPromoteEligibilityTweetFields%22%3Afalse%2C%22withBirdwatchNotes%22%3Afalse%2C%22withSuperFollowsUserFields%22%3Afalse%2C%22withDownvotePerspective%22%3Afalse%2C%22withReactionsMetadata%22%3Afalse%2C%22withReactionsPerspective%22%3Afalse%2C%22withSuperFollowsTweetFields%22%3Afalse%2C%22withVoice%22%3Afalse%2C%22withV2Timeline%22%3Atrue%7D%26features%3D%7B%22responsive_web_twitter_blue_verified_badge_is_enabled%22%3Atrue%2C%22verified_phone_label_enabled%22%3Afalse%2C%22responsive_web_graphql_timeline_navigation_enabled%22%3Atrue%2C%22unified_cards_ad_metadata_container_dynamic_card_content_query_enabled%22%3Atrue%2C%22tweetypie_unmention_optimization_enabled%22%3Atrue%2C%22responsive_web_uc_gql_enabled%22%3Atrue%2C%22vibe_api_enabled%22%3Atrue%2C%22responsive_web_edit_tweet_api_enabled%22%3Atrue%2C%22graphql_is_translatable_rweb_tweet_is_translatable_enabled%22%3Atrue%2C%22standardized_nudges_misinfo%22%3Atrue%2C%22tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled%22%3Afalse%2C%22interactive_text_enabled%22%3Atrue%2C%22responsive_web_text_conversations_enabled%22%3Afalse%2C%22responsive_web_enhance_cards_enabled%22%3Atrue%7D")
	if err != nil {
		return tweets, err
	}

	var firstjsn timelinerecursive
	err = s.RequestAPI(req, &firstjsn)
	if err != nil {
		return tweets, err
	}

	if len(firstjsn.Errors) > 0 {
		if firstjsn.Errors[0].Code != 37 {
			return tweets, fmt.Errorf("%s", firstjsn.Errors[0].Message)
		}
	}

	if len(firstjsn.Data.ThreadedConvo.Instructions) < 1 {
		return tweets, fmt.Errorf("no Instruction")
	}

	var aaatwts []contents

	cvi := 0

	for i, inst := range firstjsn.Data.ThreadedConvo.Instructions {
		if inst.Type == "TimelineAddEntries" {
			cvi = i
			break
		}
	}

	for _, entry := range firstjsn.Data.ThreadedConvo.Instructions[cvi].Entries {
		if strings.HasPrefix(entry.EntryId, "conversationthread") {

			for _, item := range entry.Content.Items {
				aaatwts = append(aaatwts, contents{
					ItemContent: item.Item.ItemContent,
					Entry:       item.EntryId,
				})
			}
		} else {
			aaatwts = append(aaatwts, contents{
				ItemContent: entry.Content.ItemContent,
				Entry:       entry.EntryId,
			})
		}
	}

	tlContents = append(aaatwts, tlContents...)

	for {
		for _, entry := range firstjsn.Data.ThreadedConvo.Instructions[cvi].Entries {
			if strings.HasPrefix(entry.EntryId, "cursor-bottom") {
				cursorBottom = entry.Content.ItemContent.Value
				break
			}
		}

		if cursorBottom != "" {
			req, err = s.newRequest("GET", "https://twitter.com/i/api/graphql/BoHLKeBvibdYDiJON1oqTg/TweetDetail?variables%3D%7B%22focalTweetId%22%3A%22"+id+"%22%2C%22cursor%22%3A%22"+cursorBottom+"%22%2C%22referrer%22%3A%22messages%22%2C%22with_rux_injections%22%3Afalse%2C%22includePromotedContent%22%3Afa%3Bse%2C%22withCommunity%22%3Afalse%2C%22withQuickPromoteEligibilityTweetFields%22%3Afalse%2C%22withBirdwatchNotes%22%3Afalse%2C%22withSuperFollowsUserFields%22%3Afalse%2C%22withDownvotePerspective%22%3Afalse%2C%22withReactionsMetadata%22%3Afalse%2C%22withReactionsPerspective%22%3Afalse%2C%22withSuperFollowsTweetFields%22%3Afalse%2C%22withVoice%22%3Atrue%2C%22withV2Timeline%22%3Atrue%7D%26features%3D%7B%22responsive_web_twitter_blue_verified_badge_is_enabled%22%3Atrue%2C%22verified_phone_label_enabled%22%3Afalse%2C%22responsive_web_graphql_timeline_navigation_enabled%22%3Atrue%2C%22unified_cards_ad_metadata_container_dynamic_card_content_query_enabled%22%3Atrue%2C%22tweetypie_unmention_optimization_enabled%22%3Atrue%2C%22responsive_web_uc_gql_enabled%22%3Atrue%2C%22vibe_api_enabled%22%3Atrue%2C%22responsive_web_edit_tweet_api_enabled%22%3Atrue%2C%22graphql_is_translatable_rweb_tweet_is_translatable_enabled%22%3Atrue%2C%22standardized_nudges_misinfo%22%3Atrue%2C%22tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled%22%3Afalse%2C%22interactive_text_enabled%22%3Atrue%2C%22responsive_web_text_conversations_enabled%22%3Afalse%2C%22responsive_web_enhance_cards_enabled%22%3Atrue%7D")
			if err != nil {
				return tweets, users, err
			}

			var backmatterjsn timelinerecursive
			err = s.RequestAPI(req, &backmatterjsn)
			if err != nil {
				return tweets, users, err
			}

			if len(backmatterjsn.Errors) > 0 {
				if backmatterjsn.Errors[0].Code != 37 {
					return tweets, users, fmt.Errorf("%s", backmatterjsn.Errors[0].Message)
				}
			}

			if len(backmatterjsn.Data.ThreadedConvo.Instructions) < 1 {
				return tweets, users, fmt.Errorf("no instructions")
			}

			convoint := 0

			for i, inst := range backmatterjsn.Data.ThreadedConvo.Instructions {
				if inst.Type == "TimelineAddEntries" {
					convoint = i
					break
				}
			}

			var inttwts []contents

			for _, entry := range backmatterjsn.Data.ThreadedConvo.Instructions[convoint].Entries {
				if strings.HasPrefix(entry.EntryId, "conversationthread") {

					for _, item := range entry.Content.Items {
						inttwts = append(inttwts, contents{
							ItemContent: item.Item.ItemContent,
							Entry:       item.EntryId,
						})
					}
				} else {
					inttwts = append(inttwts, contents{
						ItemContent: entry.Content.ItemContent,
						Entry:       entry.EntryId,
					})
				}
			}

			tlContents = append(inttwts, tlContents...)
		} else {
			break
		}
	}
}

func ItemContentToTweet(content itemcontent) (Tweet, error) {
	if content.TweetResults.Result.RestId == "" {
		return Tweet{}, fmt.Errorf("no rest ID")
	}

	var hashtags []string
	for _, ht := range content.TweetResults.Result.Legacy.Entities.Hashtags {
		hashtags = append(hashtags, ht.Text)
	}

	timestamp, err := strconv.ParseInt(content.TweetResults.Result.Legacy.CreatedAt, 10, 64)
	if err != nil {
		return Tweet{}, err
	}

	tweetid := content.TweetResults.Result.EditControl.InitialTweetId
	if tweetid == "" {
		tweetid = content.TweetResults.Result.RestId
	}

	var mention map[string]string

	for _, m := range content.TweetResults.Result.Legacy.Entities.UserMentions {
		mention[m.IdStr] = m.Name
	}

	tweet := Tweet{
		ID:               tweetid,
		EditIds:          content.TweetResults.Result.EditControl.EditIds,
		Hashtags:         hashtags,
		ReplyingTo:       content.TweetResults.Result.Legacy.InReplyToStatusIDStr,
		QuoteRetweetId:   content.TweetResults.Result.Legacy.QuotedStatusIDStr,
		IsReply:          content.TweetResults.Result.Legacy.InReplyToStatusIDStr == "",
		IsRetweet:        false,
		PermanentURL:     "https://twitter.com/i/status" + content.TweetResults.Result.RestId,
		Media:            content.TweetResults.Result.Legacy.ExtendedEntities.Media,
		Card:             content.TweetResults.Result.Card,
		Mentions:         mention,
		Text:             content.TweetResults.Result.Legacy.FullText,
		TimeParsed:       time.Now(),
		Timestamp:        timestamp,
		UserID:           content.TweetResults.Result.Legacy.UserIDStr,
		Username:         content.TweetResults.Result.Core.UserResults.Result.Legacy.Name,
		SensitiveContent: content.TweetResults.Result.Legacy.PossiblySensitive,
		Likes:            content.TweetResults.Result.Legacy.FavoriteCount,
		Retweets:         content.TweetResults.Result.Legacy.RetweetCount,
		QuoteRetweets:    content.TweetResults.Result.Legacy.QuoteCount,
		Replies:          content.TweetResults.Result.Legacy.ReplyCount,
		IsTombstone:      false,
	}

	return tweet, nil
}
