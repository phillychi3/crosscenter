package sites

import (
	"bytes"
	"crosscenter/core"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/peterbourgon/diskv/v3"
	"github.com/tidwall/gjson"
)

type BlueSkyPoster struct{}

func (dp BlueSkyPoster) Post(post PostInterface, setting core.SettingYaml, db *diskv.Diskv) (string, error) {
	return PostBlueSky(post, setting)
}

type BSKYPost struct {
	Author  string
	Content string
	Url     string
	Images  []string
	Data    uint64
	Id      string
}

type FeedResponse struct {
	Feed []BskyFeedPost `json:"feed"`
}

type BskyFeedPost struct {
	Post struct {
		Uri       string     `json:"uri"`
		Cid       string     `json:"cid"`
		Author    BskyAuthor `json:"author"`
		Record    BskyRecord `json:"record"`
		IndexedAt time.Time  `json:"indexedAt"`
		Embed     BskyEmbed  `json:"embed"`
	} `json:"post"`
}

type BskyImage struct {
	Thumb       string `json:"thumb"`
	Fullsize    string `json:"fullsize"`
	Alt         string `json:"alt"`
	AspectRatio struct {
		Height int `json:"height"`
		Width  int `json:"width"`
	} `json:"aspectRatio"`
}

type BskyEmbed struct {
	Images []BskyImage `json:"images"`
}

type BskyAuthor struct {
	Did         string `json:"did"`
	Handle      string `json:"handle"`
	DisplayName string `json:"displayName"`
}

type BskyRecord struct {
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"createdAt"`
}

type CreateSessionResponse struct {
	AccessJwt  string    `json:"accessJwt"`
	RefreshJwt string    `json:"refreshJwt"`
	Handle     string    `json:"handle"`
	Did        string    `json:"did"`
	CreatedAt  time.Time `json:"createdAt"`
}

func (t BSKYPost) GetAuthor() string   { return t.Author }
func (t BSKYPost) GetContent() string  { return t.Content }
func (t BSKYPost) GetURL() string      { return t.Url }
func (t BSKYPost) GetImages() []string { return t.Images }
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
			Author:  post.Author.Handle,
			Content: post.Record.Text,
			Url:     fmt.Sprintf("https://bsky.app/profile/%s/post/%s", post.Author.Handle, strings.Split(post.Uri, "/")[len(strings.Split(post.Uri, "/"))-1]),
			Images:  images,
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

func createBskySession(setting core.SettingYaml) (*CreateSessionResponse, error) {

	payload := map[string]string{
		"identifier": setting.BlueSky.DID,
		"password":   setting.BlueSky.Password,
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("error marshalling payload to JSON: %w", err)
	}
	req, err := http.NewRequest(
		"POST",
		"https://bsky.social/xrpc/com.atproto.server.createSession",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
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

	var session CreateSessionResponse
	if err := json.Unmarshal(body, &session); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	return &session, nil
}

type ref struct {
	Link string `json:"$link"`
}

type blobResponse struct {
	Blob blob `json:"blob"`
}

type blob struct {
	Ref      ref    `json:"ref"`
	MimeType string `json:"mimeType"`
	Size     int    `json:"size"`
}

func postBlueskyBlob(image []byte, session *CreateSessionResponse) (*blobResponse, error) {
	core.Debug("Posting blob to BlueSky")
	url := "https://bsky.social/xrpc/com.atproto.repo.uploadBlob"

	req, err := http.NewRequest(
		"POST",
		url,
		bytes.NewBuffer(image),
	)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "image/png")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", session.AccessJwt))

	client := &http.Client{Timeout: 10 * time.Second}
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

	var blob blobResponse

	err = json.Unmarshal(body, &blob)
	if err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}
	return &blob, nil
}

func PostBlueSky(post PostInterface, setting core.SettingYaml) (string, error) {
	core.Debug("Posting to BlueSky")
	session, err := createBskySession(setting)
	if err != nil {
		return "", err
	}

	url := "https://bsky.social/xrpc/com.atproto.repo.createRecord"

	// record {
	//     "$type": "app.bsky.feed.post",
	//     "text": "example post with multiple images attached",
	//     "createdAt": "2023-08-07T05:49:35.422015Z",
	//     "embed": {
	//       "$type": "app.bsky.embed.images",
	//       "images": [
	//         {
	//           "alt": "brief alt text description of the first image",
	//           "image": {
	//             "$type": "blob",
	//             "ref": {
	//               "$link": "bafkreibabalobzn6cd366ukcsjycp4yymjymgfxcv6xczmlgpemzkz3cfa"
	//             },
	//             "mimeType": "image/webp",
	//             "size": 760898
	//           }
	//         },
	//         {
	//           "alt": "brief alt text description of the second image",
	//           "image": {
	//             "$type": "blob",
	//             "ref": {
	//               "$link": "bafkreif3fouono2i3fmm5moqypwskh3yjtp7snd5hfq5pr453oggygyrte"
	//             },
	//             "mimeType": "image/png",
	//             "size": 13208
	//           }
	//         }
	//       ]
	//     }
	//   }
	var record map[string]any
	if len(post.GetImages()) > 0 {
		core.Debug("Posting with images")
		images := []map[string]any{}
		for _, image := range post.GetImages() {
			imgBytes, err := core.GetImageBytes(image)
			if err != nil {
				return "", err
			}
			ref, err := postBlueskyBlob(imgBytes, session)
			if err != nil {
				return "", err
			}
			images = append(images, map[string]any{
				"alt": "Image",
				"image": map[string]any{
					"$type":    "blob",
					"ref":      ref.Blob.Ref,
					"mimeType": ref.Blob.MimeType,
					"size":     ref.Blob.Size,
				},
			})
		}
		record = map[string]any{
			"$type":     "app.bsky.feed.post",
			"text":      core.TextFormat(setting.BlueSky.PostText, post),
			"createdAt": time.Now().Format(time.RFC3339),
			"embed": map[string]any{
				"$type":  "app.bsky.embed.images",
				"images": images,
			},
		}
	} else {
		record = map[string]any{
			"$type":     "app.bsky.feed.post",
			"text":      core.TextFormat(setting.BlueSky.PostText, post),
			"createdAt": time.Now().Format(time.RFC3339),
		}
	}

	payload := map[string]any{
		"repo":       setting.BlueSky.DID,
		"collection": "app.bsky.feed.post",
		"record":     record,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("error marshalling payload: %w", err)
	}

	req, err := http.NewRequest(
		"POST",
		url,
		bytes.NewBuffer(payloadBytes),
	)
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", session.AccessJwt))

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("server returned error status %d: %s", resp.StatusCode, string(body))
	}

	cid := gjson.Get(string(body), "cid").String()

	return cid, nil
}
