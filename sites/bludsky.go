package sites

import (
	"crosscenter/core"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type BSKYPost struct {
	author  string
	content string
	url     string
	images  []string
	Data    uint64
	Id      string
}

type FeedResponse struct {
	Feed []Post `json:"feed"`
}

type Post struct {
	Post struct {
		Uri       string    `json:"uri"`
		Cid       string    `json:"cid"`
		Author    Author    `json:"author"`
		Record    Record    `json:"record"`
		IndexedAt time.Time `json:"indexedAt"`
		Embed     Embed     `json:"embed"`
	} `json:"post"`
}

type Image struct {
	Thumb       string `json:"thumb"`
	Fullsize    string `json:"fullsize"`
	Alt         string `json:"alt"`
	AspectRatio struct {
		Height int `json:"height"`
		Width  int `json:"width"`
	} `json:"aspectRatio"`
}

type Embed struct {
	Images []Image `json:"images"`
}

type Author struct {
	Did         string `json:"did"`
	Handle      string `json:"handle"`
	DisplayName string `json:"displayName"`
}

type Record struct {
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"createdAt"`
}

func (t BSKYPost) GetAuthor() string   { return t.author }
func (t BSKYPost) GetContent() string  { return t.content }
func (t BSKYPost) GetURL() string      { return t.url }
func (t BSKYPost) GetImages() []string { return t.images }
func (t BSKYPost) GetDate() uint64     { return t.Data }
func (t BSKYPost) GetID() string       { return t.Id }

func GetBSKY(setting core.SettingYaml) ([]PostInterface, error) {
	did := setting.BlueSky.DID

	feed, err := getAuthorFeed(did)
	if err != nil {
		fmt.Printf("Error getting feed: %v\n", err)
		return nil, err
	}
	var posts []PostInterface

	for _, item := range feed.Feed {
		post := item.Post
		images := []string{}
		for _, image := range post.Embed.Images {
			images = append(images, image.Fullsize)
		}
		bpost := BSKYPost{
			author:  post.Author.Handle,
			content: post.Record.Text,
			url:     fmt.Sprintf("https://bsky.app/profile/%s/post/%s", post.Author.Handle, strings.Split(post.Uri, "/")[len(strings.Split(post.Uri, "/"))-1]),
			images:  images,
			Data:    uint64(post.Record.CreatedAt.Unix()),
			Id:      post.Cid,
		}
		posts = append(posts, bpost)
	}
	return posts, nil

}

func getAuthorFeed(did string) (*FeedResponse, error) {
	url := fmt.Sprintf("https://public.api.bsky.app/xrpc/app.bsky.feed.getAuthorFeed?actor=%s&filter=posts_and_author_threads&includePins=true&limit=30", did)
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned error status %d: %s", resp.StatusCode, string(body))
	}

	var feed FeedResponse
	if err := json.Unmarshal(body, &feed); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	return &feed, nil
}
