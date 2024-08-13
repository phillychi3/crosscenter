package sites

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

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
	context string
	url     string
	images  []string
	Data    uint64
}

type Threadsuser struct {
	Username string
}

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

func GetThreadsUserId(threadsuser Threadsuser, lsdtoken Tokens) (string, error) {
	lsd := lsdtoken.LSD
	pathName := fmt.Sprintf("/@%s", threadsuser.Username)
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

	headers := ThreadHeader(threadsuser.Username, lsd)
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

func GetThreadsPosts(threadsuser Threadsuser) ([]ThreadsPost, error) {
	// 	curl --request POST \
	//   --url https://www.threads.net/api/graphql \
	//   --header 'user-agent: threads-client' \
	//   --header 'x-ig-app-id: 238260118697367' \
	//   --header 'content-type: application/x-www-form-urlencoded' \
	//   --data 'variables={"userID":"314216"}' \
	//   --data doc_id=6232751443445612
	tokens, err := getToken(threadsuser.Username)
	if err != nil {
		return nil, err
	}
	threadsUserId, err := GetThreadsUserId(threadsuser, *tokens)
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

	headers := ThreadHeader(threadsuser.Username, tokens.LSD)
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
		posts := thread.Get(fmt.Sprintf(`thread_items.#(post.user.username=="%s")#`, threadsuser.Username))

		posts.ForEach(func(_, post gjson.Result) bool {
			threadpost := ThreadsPost{
				author:  post.Get("post.user.username").String(),
				context: post.Get("post.caption.text").String(),
				url:     post.Get("post.code").String(),
				Data:    post.Get("post.taken_at").Uint(),
				images:  []string{},
			}
			threadposts = append(threadposts, threadpost)
			return true
		})
		return true
	})

	return threadposts, nil
}
