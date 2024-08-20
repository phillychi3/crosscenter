package sites

import (
	"crosscenter/core"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/peterbourgon/diskv/v3"
	"github.com/tidwall/gjson"
	"golang.org/x/net/html"
)

const (
	ApiUrl     = "https://www.threads.net/api/graphql"
	appId      = "238260118697367"
	USER_AGENT = "Barcelona 289.0.0.77.109 Android"
	asbdId     = "129477"
)

type ThreadsPost struct {
	author  string
	content string
	url     string
	images  []string
	Data    uint64
	id      string
}

func (t ThreadsPost) GetAuthor() string   { return t.author }
func (t ThreadsPost) GetContent() string  { return t.content }
func (t ThreadsPost) GetURL() string      { return t.url }
func (t ThreadsPost) GetImages() []string { return t.images }
func (t ThreadsPost) GetData() uint64     { return t.Data }
func (t ThreadsPost) GetId() string       { return t.id }

type Tokens struct {
	LSD string
}

// https://github.com/DIYgod/RSSHub/blob/master/lib/routes/threads/utils.ts#19
func getToken(user string) (*Tokens, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", "https://www.threads.net/@"+user, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", USER_AGENT)
	req.Header.Set("X-IG-App-ID", appId)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}

	var scriptContent string
	var findScript func(*html.Node)
	findScript = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "script" {
			for _, a := range n.Attr {
				if a.Key == "type" && a.Val == "application/json" && strings.Contains(n.FirstChild.Data, "LSD") {
					scriptContent = n.FirstChild.Data
					return
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findScript(c)
		}
	}
	findScript(doc)

	if scriptContent == "" {
		return nil, errors.New("LSD token not found, html parsing error")
	}
	re := regexp.MustCompile(`"LSD",\[\],{"token":"([\w@-]+)"},`)
	matches := re.FindStringSubmatch(scriptContent)
	if len(matches) < 2 {
		return nil, errors.New("LSD token not found")
	}

	lsd := matches[1]
	return &Tokens{LSD: lsd}, nil
}

func ThreadHeader(user string, lsd string) map[string]string {
	return map[string]string{
		"Accept":         "*/*",
		"Host":           "www.threads.net",
		"Origin":         "https://www.threads.net",
		"Referer":        "https://www.threads.net/@" + user,
		"User-Agent":     USER_AGENT,
		"X-IG-App-ID":    appId,
		"X-FB-LSD":       lsd,
		"Sec-Fetch-Site": "same-origin",
	}
}

func GetThreadsUserId(username string, lsdtoken Tokens) (string, error) {
	lsd := lsdtoken.LSD
	pathName := fmt.Sprintf("/@%s", username)
	payload := url.Values{
		"route_urls[0]":     {pathName},
		"routing_namespace": {"barcelona_web"},
		"__user":            {"0"},
		"__a":               {"1"},
		"__comet_req":       {"29"},
		"lsd":               {lsd},
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", "https://www.threads.net/ajax/bulk-route-definitions/", strings.NewReader(payload.Encode()))
	if err != nil {
		return "", err
	}

	headers := ThreadHeader(username, lsd)
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-ASBD-ID", asbdId)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	jsonStr := string(body)[9:]
	var result map[string]interface{}
	err = json.Unmarshal([]byte(jsonStr), &result)
	if err != nil {
		return "", fmt.Errorf("JSON 解析錯誤: %v", err)
	}

	if errMsg, ok := result["error"].(float64); ok {
		return "", fmt.Errorf("API 錯誤代碼: %f", errMsg)
	}

	if errDesc, ok := result["errorDescription"].(string); ok {
		return "", fmt.Errorf("API 錯誤描述: %s", errDesc)
	}
	userId := result["payload"].(map[string]interface{})["payloads"].(map[string]interface{})[pathName].(map[string]interface{})["result"].(map[string]interface{})["exports"].(map[string]interface{})["rootView"].(map[string]interface{})["props"].(map[string]interface{})["user_id"].(string)
	return userId, nil
}

func GetThreadsPosts(setting core.SettingYaml) ([]ThreadsPost, error) {
	// 	curl --request POST \
	//   --url https://www.threads.net/api/graphql \
	//   --header 'user-agent: threads-client' \
	//   --header 'x-ig-app-id: 238260118697367' \
	//   --header 'content-type: application/x-www-form-urlencoded' \
	//   --data 'variables={"userID":"314216"}' \
	//   --data doc_id=6232751443445612
	tokens, err := getToken(setting.Threads.Username)
	if err != nil {
		return nil, err
	}
	threadsUserId, err := GetThreadsUserId(setting.Threads.Username, *tokens)
	if err != nil {
		return nil, err
	}

	variables := map[string]string{"userID": threadsUserId}
	variablesJSON, err := json.Marshal(variables)
	if err != nil {
		return nil, err
	}

	payload := url.Values{
		"variables": {string(variablesJSON)},
		"doc_id":    {"6232751443445612"},
		"lsd":       {tokens.LSD},
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", ApiUrl, strings.NewReader(payload.Encode()))

	headers := ThreadHeader(setting.Threads.Username, tokens.LSD)
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if err != nil {
		return nil, err
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

	var result interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}
	threads := gjson.Get(string(body), "data.mediaData.threads")
	if !threads.Exists() {
		return nil, errors.New("threads not found")
	}

	var threadposts []ThreadsPost

	threads.ForEach(func(_, thread gjson.Result) bool {
		posts := thread.Get(fmt.Sprintf(`thread_items.#(post.user.username=="%s")#`, setting.Threads.Username))

		posts.ForEach(func(_, post gjson.Result) bool {
			postData := post.Get("post")

			var images []string
			if postData.Get("carousel_media_count").Type != gjson.Null {
				postData.Get("carousel_media").ForEach(func(_, media gjson.Result) bool {
					images = append(images, media.Get("image_versions2.candidates.0.url").String())
					return true
				})
			} else {
				images = append(images, postData.Get("image_versions2.candidates.0.url").String())
			}

			threadpost := ThreadsPost{
				author:  postData.Get("user.username").String(),
				content: postData.Get("caption.text").String(),
				url:     "https://www.threads.net/@" + setting.Threads.Username + "/post/" + postData.Get("code").String(),
				Data:    postData.Get("taken_at").Uint(),
				images:  images,
				id:      postData.Get("code").String(),
			}

			threadposts = append(threadposts, threadpost)
			return true
		})
		return true
	})

	return threadposts, nil
}

func createThreadsSingleMediaContainer(post PostInterface, setting core.SettingYaml, db *diskv.Diskv) (string, error) {
	containerurl := fmt.Sprintf("https://graph.threads.net/v1.0/%s/threads", setting.Threads.Username)
	access_token, err := db.Read("threads_access_token")
	if err != nil {
		return "", err
	}
	payload := url.Values{
		"media_type":   {"TEXT"},
		"text":         {"Hello, World!"},
		"access_token": {string(access_token)},
	}
	// payload := url.Values{
	// 	"media_type":   {"TEXT"},
	// 	"text":         {post.GetContent()},
	// 	"access_token": {string(access_token)},
	// }
	client := &http.Client{}
	req, err := http.NewRequest("POST", containerurl, strings.NewReader(payload.Encode()))
	if err != nil {
		return "", err
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
	id := gjson.Get(string(body), "id").String()
	fmt.Println(id)
	return id, nil

}

func createThreadsCarouselContainer(post PostInterface, setting core.SettingYaml, mediaContainers []string, db *diskv.Diskv) (string, error) {
	containerurl := fmt.Sprintf("https://graph.threads.net/v1.0/%s/threads", setting.Threads.Username)
	access_token, err := db.Read("threads_access_token")
	if err != nil {
		return "", err
	}
	payload := url.Values{
		"media_type":   {"CAROUSEL"},
		"children":     {strings.Join(mediaContainers, ",")},
		"access_token": {string(access_token)},
	}
	client := &http.Client{}
	req, err := http.NewRequest("POST", containerurl, strings.NewReader(payload.Encode()))
	if err != nil {
		return "", err
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
	id := gjson.Get(string(body), "id").String()
	fmt.Println(id)
	return id, nil
}

func reflashaccesstoken(setting core.SettingYaml, db *diskv.Diskv) error {
	url := fmt.Sprintf("https://graph.threads.net/refresh_access_token?grant_type=th_refresh_token&access_token=%s", setting.Threads.AccessToken)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	access_token := gjson.Get(string(body), "access_token").String()
	db.Write("threads_access_token", []byte(access_token))
	return nil
}

func getlongaccesstoken(setting core.SettingYaml, db *diskv.Diskv) error {
	url := fmt.Sprintf("https://graph.threads.net/access_token?grant_type=th_exchange_token&client_secret=%s&access_token=%s", setting.Threads.ClientSecret, setting.Threads.AccessToken)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	access_token := gjson.Get(string(body), "access_token").String()
	db.Write("threads_access_token", []byte(access_token))
	return nil
}

type ThreadPoster struct{}

func (tp ThreadPoster) Post(post PostInterface, setting core.SettingYaml, db *diskv.Diskv) error {
	return SendThreadPost(post, setting, db)
}

func SendThreadPost(post PostInterface, setting core.SettingYaml, db *diskv.Diskv) error {
	sendurl := fmt.Sprintf("https://graph.threads.net/v1.0/%s/threads_publish", setting.Threads.Username)
	access_token, err := db.Read("threads_access_token")
	if err != nil {
		err = getlongaccesstoken(setting, db)
		if err != nil {
			return err
		}
	}
	postid, err := createThreadsSingleMediaContainer(post, setting, db)
	if err != nil {
		return err
	}
	payload := url.Values{
		"creation_id":  {postid},
		"access_token": {string(access_token)},
	}
	client := &http.Client{}
	req, err := http.NewRequest("POST", sendurl, strings.NewReader(payload.Encode()))
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
