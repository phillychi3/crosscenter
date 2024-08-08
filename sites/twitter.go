package sites

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/tidwall/gjson"
)

// 類型 twitter

type Twitteruser struct {
	Username string
	Token    string
}

func getGuestToken() (string, error) {
	url := "https://api.twitter.com/1.1/guest/activate.json"

	client := &http.Client{}
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return "", err
	}

	req.Header = http.Header{
		"User-Agent":                {"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:93.0) Gecko/20100101 Firefox/93.0"},
		"Accept":                    {"*/*"},
		"Accept-Language":           {"zh-CN,zh;q=0.8,zh-TW;q=0.7,zh-HK;q=0.5,en-US;q=0.3,en;q=0.2"},
		"x-guest-token":             {""},
		"x-twitter-client-language": {"zh-cn"},
		"x-twitter-active-user":     {"yes"},
		"x-csrf-token":              {"25ea9d09196a6ba850201d47d7e75733"},
		"Sec-Fetch-Dest":            {"empty"},
		"Sec-Fetch-Mode":            {"cors"},
		"Sec-Fetch-Site":            {"same-origin"},
		"authorization":             {"Bearer AAAAAAAAAAAAAAAAAAAAANRILgAAAAAAnNwIzUejRCOuH5E6I8xnZz4puTs%3D1Zv7ttfk8LF81IUq16cHjhLTvJu4FA33AGWWjCpTnA"},
		"Referer":                   {"https://twitter.com/"},
		"Connection":                {"keep-alive"},
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	guestToken := gjson.GetBytes(body, "guest_token").String()
	return guestToken, nil
}

func getTwitterUserId(name string) (string, error) {

	UserByScreenName := "https://x.com/i/api/graphql/sLVLhk0bGj3MVFEKTdax1w/UserByScreenName?variables=%s&features=%s"
	guesttoken, err := getGuestToken()
	if err != nil {
		return "", err
	}
	header := map[string]string{
		"Content-Type":  "application/json",
		"x-guest-token": guesttoken,
		"authorization": "Bearer AAAAAAAAAAAAAAAAAAAAANRILgAAAAAAnNwIzUejRCOuH5E6I8xnZz4puTs%3D1Zv7ttfk8LF81IUq16cHjhLTvJu4FA33AGWWjCpTnA",
		"x-csrf-token":  "25ea9d09196a6ba850201d47d7e75733",
	}

	variables := map[string]interface{}{
		"screen_name":              name,
		"withSafetyModeUserFields": true,
	}

	features := map[string]bool{
		"blue_business_profile_image_shape_enabled":                         true,
		"hidden_profile_likes_enabled":                                      true,
		"hidden_profile_subscriptions_enabled":                              true,
		"responsive_web_graphql_exclude_directive_enabled":                  true,
		"verified_phone_label_enabled":                                      false,
		"subscriptions_verification_info_is_identity_verified_enabled":      true,
		"subscriptions_verification_info_verified_since_enabled":            true,
		"highlights_tweets_tab_ui_enabled":                                  true,
		"creator_subscriptions_tweet_preview_api_enabled":                   true,
		"responsive_web_graphql_skip_user_profile_image_extensions_enabled": false,
		"responsive_web_graphql_timeline_navigation_enabled":                true,
	}

	variablesJSON, _ := json.Marshal(variables)
	featuresJSON, _ := json.Marshal(features)

	reqURL := fmt.Sprintf(UserByScreenName, url.QueryEscape(string(variablesJSON)), url.QueryEscape(string(featuresJSON)))

	client := &http.Client{}
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return "", err
	}

	for key, value := range header {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	fmt.Println(string(body))

	userId := gjson.GetBytes(body, "data.user.result.rest_id").String()
	return userId, nil
}

func GetTwitterPosts(twitter Twitteruser) ([]map[string]string, error) {
	guesttoken, err := getGuestToken()
	if err != nil {
		return nil, err
	}

	useridurl := "https://x.com/i/api/graphql/V7H0Ap3_Hh2FyS75OCDO3Q/UserTweets?variables=%s&features=%s"

	FEATURES := map[string]bool{
		"rweb_tipjar_consumption_enabled":                                         true,
		"responsive_web_graphql_exclude_directive_enabled":                        true,
		"verified_phone_label_enabled":                                            false,
		"creator_subscriptions_tweet_preview_api_enabled":                         true,
		"responsive_web_graphql_timeline_navigation_enabled":                      true,
		"responsive_web_graphql_skip_user_profile_image_extensions_enabled":       false,
		"communities_web_enable_tweet_community_results_fetch":                    true,
		"c9s_tweet_anatomy_moderator_badge_enabled":                               true,
		"articles_preview_enabled":                                                true,
		"tweetypie_unmention_optimization_enabled":                                true,
		"responsive_web_edit_tweet_api_enabled":                                   true,
		"graphql_is_translatable_rweb_tweet_is_translatable_enabled":              true,
		"view_counts_everywhere_api_enabled":                                      true,
		"longform_notetweets_consumption_enabled":                                 true,
		"responsive_web_twitter_article_tweet_consumption_enabled":                true,
		"tweet_awards_web_tipping_enabled":                                        false,
		"creator_subscriptions_quote_tweet_preview_enabled":                       false,
		"freedom_of_speech_not_reach_fetch_enabled":                               true,
		"standardized_nudges_misinfo":                                             true,
		"tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled": true,
		"rweb_video_timestamps_enabled":                                           true,
		"longform_notetweets_rich_text_read_enabled":                              true,
		"longform_notetweets_inline_media_enabled":                                true,
		"responsive_web_enhance_cards_enabled":                                    false,
	}
	header := map[string]string{
		"Content-Type":  "application/json",
		"x-guest-token": guesttoken,
		"authorization": "Bearer AAAAAAAAAAAAAAAAAAAAANRILgAAAAAAnNwIzUejRCOuH5E6I8xnZz4puTs%3D1Zv7ttfk8LF81IUq16cHjhLTvJu4FA33AGWWjCpTnA",
		"x-csrf-token":  "25ea9d09196a6ba850201d47d7e75733",
	}

	userid, err := getTwitterUserId(twitter.Username)
	if err != nil {
		return nil, err
	}

	fmt.Println(userid)

	variables := map[string]interface{}{
		"userId":                 userid,
		"count":                  20,
		"withHighlightedLabel":   true,
		"withTweetQuoteCount":    true,
		"includePromotedContent": true,
		"withTweetResult":        false,
		"withReactions":          false,
		"withUserResults":        false,
		"withVoice":              false,
		"withNonLegacyCard":      true,
		"withBirdwatchPivots":    false,
		"cursor":                 "CURSOR",
	}

	variablesJSON, _ := json.Marshal(variables)
	featuresJSON, _ := json.Marshal(FEATURES)

	reqURL := fmt.Sprintf(useridurl, url.QueryEscape(string(variablesJSON)), url.QueryEscape(string(featuresJSON)))

	client := &http.Client{}
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, err
	}

	for key, value := range header {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	jsonResult := gjson.ParseBytes(body)

	fmt.Println(jsonResult)

	if !jsonResult.Get("data.user").Exists() {
		return nil, fmt.Errorf("cannot find user")
	}

	instructions := jsonResult.Get("data.user.result.timeline.timeline.instructions")
	if instructions.Get("#").Int() < 3 {
		return nil, fmt.Errorf("cannot load user timeline")
	}

	var listOfPosts []map[string]string

	entries := instructions.Get("2.entries")
	entries.ForEach(func(key, value gjson.Result) bool {
		fullText := value.Get("content.itemContent.tweet_results.result.legacy.full_text").String()
		listOfPosts = append(listOfPosts, map[string]string{"context": fullText})
		return true
	})

	return listOfPosts, nil
}

func PostTwitterPost() {

}
