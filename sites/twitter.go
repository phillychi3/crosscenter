package sites

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"time"

	"crosscenter/core"

	"github.com/dghubble/oauth1"
	"github.com/peterbourgon/diskv/v3"
	"github.com/tidwall/gjson"
)

// 類型 twitter

type TwitterPost struct {
	Author    string
	Author_id string
	Content   string
	Url       string
	Images    []string
	Data      uint64
	Id        string
}

func (t TwitterPost) GetAuthor() string   { return t.Author }
func (t TwitterPost) GetContent() string  { return t.Content }
func (t TwitterPost) GetURL() string      { return t.Url }
func (t TwitterPost) GetImages() []string { return t.Images }
func (t TwitterPost) GetDate() uint64     { return t.Data }
func (t TwitterPost) GetID() string       { return t.Id }

func getGuestToken() (string, error) {
	url := "https://api.x.com/1.1/guest/activate.json"

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
		"Referer":                   {"https://x.com/"},
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

func getTwitterUserId(name string, setting core.SettingYaml) (string, error) {

	UserByScreenName := "https://x.com/i/api/graphql/laYnJPCAcVo0o6pzcnlVxQ/UserByScreenName?variables=%s&features=%s&fieldToggles=%s"
	header := map[string]string{
		"Content-Type":  "application/json",
		"authorization": "Bearer AAAAAAAAAAAAAAAAAAAAANRILgAAAAAAnNwIzUejRCOuH5E6I8xnZz4puTs%3D1Zv7ttfk8LF81IUq16cHjhLTvJu4FA33AGWWjCpTnA",
	}

	variables := map[string]interface{}{
		"screen_name": name,
	}

	features := map[string]bool{
		"hidden_profile_subscriptions_enabled":                              true,
		"rweb_tipjar_consumption_enabled":                                   true,
		"responsive_web_graphql_exclude_directive_enabled":                  true,
		"verified_phone_label_enabled":                                      false,
		"subscriptions_verification_info_is_identity_verified_enabled":      true,
		"subscriptions_verification_info_verified_since_enabled":            true,
		"highlights_tweets_tab_ui_enabled":                                  true,
		"responsive_web_twitter_article_notes_tab_enabled":                  true,
		"subscriptions_feature_can_gift_premium":                            true,
		"creator_subscriptions_tweet_preview_api_enabled":                   true,
		"responsive_web_graphql_skip_user_profile_image_extensions_enabled": false,
		"responsive_web_graphql_timeline_navigation_enabled":                true,
	}

	filedtoggles := map[string]bool{
		"withAuxiliaryUserLabels": false,
	}

	variablesJSON, _ := json.Marshal(variables)
	featuresJSON, _ := json.Marshal(features)
	filedtogglesJSON, _ := json.Marshal(filedtoggles)

	reqURL := fmt.Sprintf(UserByScreenName, url.QueryEscape(string(variablesJSON)), url.QueryEscape(string(featuresJSON)), url.QueryEscape(string(filedtogglesJSON)))

	client := &http.Client{}
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return "", err
	}
	for key, value := range header {
		req.Header.Set(key, value)
	}
	req.Header.Set("x-Csrf-Token", setting.Twitter.Ct0)

	req.AddCookie(&http.Cookie{
		Name:  "auth_token",
		Value: setting.Twitter.Auth_token,
	})
	req.AddCookie(&http.Cookie{
		Name:  "ct0",
		Value: setting.Twitter.Ct0,
	})

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

	userId := gjson.GetBytes(body, "data.user.result.rest_id").String()
	return userId, nil
}

func GetTwitterPosts(setting core.SettingYaml) ([]PostInterface, error) {
	// 無法獲取全部貼文
	if setting.Twitter.Username == "" {
		return nil, fmt.Errorf("twitter username cannot be empty")
	}
	guesttoken, err := getGuestToken()
	if err != nil {
		return nil, err
	}

	useridurl := "https://x.com/i/api/graphql/Tg82Ez_kxVaJf7OPbUdbCg/UserTweets?variables=%s&features=%s"

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
		"Content-Type":           "application/json",
		"x-guest-token":          guesttoken,
		"'x-twitter-active-user": "yes",
		"authorization":          "Bearer AAAAAAAAAAAAAAAAAAAAANRILgAAAAAAnNwIzUejRCOuH5E6I8xnZz4puTs%3D1Zv7ttfk8LF81IUq16cHjhLTvJu4FA33AGWWjCpTnA",
		"x-csrf-token":           "25ea9d09196a6ba850201d47d7e75733",
	}

	userid, err := getTwitterUserId(setting.Twitter.Username, setting)
	if err != nil {
		return nil, err
	}

	variables := map[string]interface{}{
		"userId":                                 userid,
		"count":                                  20,
		"includePromotedContent":                 true,
		"withQuickPromoteEligibilityTweetFields": true,
		"withVoice":                              true,
		// "withV2Timeline":                         true,
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
	req.Header.Set("x-csrf-token", setting.Twitter.Ct0)

	req.AddCookie(&http.Cookie{
		Name:  "auth_token",
		Value: setting.Twitter.Auth_token,
	})
	req.AddCookie(&http.Cookie{
		Name:  "ct0",
		Value: setting.Twitter.Ct0,
	})

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

	if !jsonResult.Get("data.user").Exists() {
		return nil, fmt.Errorf("cannot find user")
	}

	instructions := jsonResult.Get("data.user.result.timeline.timeline.instructions")
	if instructions.Get("#").Int() < 3 {
		return nil, fmt.Errorf("cannot load user timeline")
	}

	var listOfPosts []PostInterface

	entries := instructions.Get("#(type=TimelineAddEntries).entries")
	//TODO: entryId profile-conversation 尚未適配 who-to-follow需要避開
	for _, value := range entries.Array() {
		//TODO: 簡易解決 : 判斷content.entryType == TimelineTimelineItem
		if value.Get("content.entryType").String() != "TimelineTimelineItem" {
			continue
		}
		t, err := time.Parse("Mon Jan 2 15:04:05 -0700 2006", value.Get("content.itemContent.tweet_results.result.legacy.created_at").String())
		if err != nil {
			t = time.Now()
		}
		images := []string{}
		value.Get("content.itemContent.tweet_results.result.legacy.entities.media").ForEach(func(_, image gjson.Result) bool {
			images = append(images, image.Get("media_url_https").String())
			return true
		})
		twitterpost := TwitterPost{
			Author:    value.Get("content.itemContent.tweet_results.result.core.user_results.result.legacy.name").String(),
			Author_id: value.Get("content.itemContent.tweet_results.result.core.user_results.result.legacy.screen_name").String(),
			Content:   value.Get("content.itemContent.tweet_results.result.legacy.full_text").String(),
			Url:       "https://x.com/" + value.Get("content.itemContent.tweet_results.result.core.user_results.result.legacy.screen_name").String() + "/status/" + value.Get("content.itemContent.tweet_results.result.rest_id").String(),
			Data:      uint64(t.Unix()),
			Images:    images,
			Id:        value.Get("content.itemContent.tweet_results.result.rest_id").String(),
		}
		listOfPosts = append(listOfPosts, twitterpost)
	}
	return listOfPosts, nil
}

type TwitterPoster struct{}

func (tp TwitterPoster) Post(post PostInterface, setting core.SettingYaml, db *diskv.Diskv) (string, error) {
	return PostTwitterPost(post, setting)
}

func uploadMediaToTwitter(image string, setting core.SettingYaml) (string, error) {
	consumerKey := setting.Twitter.CONSUMERKEY
	consumerSecret := setting.Twitter.CONSUMERSECRET
	accessToken := setting.Twitter.ACCESSTOKEN
	accessTokenSecret := setting.Twitter.ACCESSTOKENSECRET

	config := oauth1.NewConfig(consumerKey, consumerSecret)
	token := oauth1.NewToken(accessToken, accessTokenSecret)

	httpClient := config.Client(oauth1.NoContext, token)

	imageresp, err := http.Get(image)
	if err != nil {
		return "", fmt.Errorf("下載圖片失敗: %v", err)
	}
	defer imageresp.Body.Close()

	fileContent, err := io.ReadAll(imageresp.Body)
	if err != nil {
		return "", fmt.Errorf("讀取圖片失敗: %v", err)
	}

	b := &bytes.Buffer{}
	form := multipart.NewWriter(b)

	fw, err := form.CreateFormFile("media", "file.jpg")
	if err != nil {
		return "", err
	}

	if _, err := io.Copy(fw, bytes.NewReader(fileContent)); err != nil {
		return "", err
	}

	form.Close()

	uploadResp, err := httpClient.Post("https://upload.twitter.com/1.1/media/upload.json?media_category=tweet_image", form.FormDataContentType(), bytes.NewReader(b.Bytes()))
	if err != nil {
		return "", err
	}
	defer uploadResp.Body.Close()

	body, err := io.ReadAll(uploadResp.Body)
	if err != nil {
		return "", err
	}

	id := gjson.GetBytes(body, "media_id_string").String()
	if id == "" {
		return "", fmt.Errorf("無法獲取 media_id_string")
	}

	return id, nil
}

func PostTwitterPost(post PostInterface, setting core.SettingYaml) (string, error) {

	consumerKey := setting.Twitter.CONSUMERKEY
	consumerSecret := setting.Twitter.CONSUMERSECRET
	accessToken := setting.Twitter.ACCESSTOKEN
	accessTokenSecret := setting.Twitter.ACCESSTOKENSECRET

	config := oauth1.NewConfig(consumerKey, consumerSecret)
	token := oauth1.NewToken(accessToken, accessTokenSecret)

	httpClient := config.Client(oauth1.NoContext, token)
	var tweet map[string]interface{}
	if len(post.GetImages()) > 0 {
		mediaIds := []string{}
		for _, image := range post.GetImages() {
			mediaId, err := uploadMediaToTwitter(image, setting)
			if err != nil {
				fmt.Println("Error uploading media:", err)
				return "", err
			}
			mediaIds = append(mediaIds, mediaId)
		}
		tweet = map[string]interface{}{
			"text": core.TextFormat(setting.Twitter.PostText, post),
			"media": map[string]interface{}{
				"media_ids": mediaIds,
			},
		}
	} else {
		tweet = map[string]interface{}{
			"text": core.TextFormat(setting.Twitter.PostText, post),
		}
	}
	jsonStr, _ := json.Marshal(tweet)

	resp, err := httpClient.Post(
		"https://api.twitter.com/2/tweets",
		"application/json",
		bytes.NewBuffer(jsonStr),
	)
	if err != nil {
		return "", fmt.Errorf("error sending tweet: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	id := gjson.GetBytes(body, "data.id").String()

	return id, nil
}
