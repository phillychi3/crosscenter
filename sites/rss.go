package sites

import (
	"crosscenter/core"

	"github.com/mmcdole/gofeed"
)

type RSSPost struct {
	author  string
	content string
	url     string
	images  []string
	Data    uint64
}

func (t RSSPost) GetAuthor() string   { return t.author }
func (t RSSPost) GetContent() string  { return t.content }
func (t RSSPost) GetURL() string      { return t.url }
func (t RSSPost) GetImages() []string { return t.images }
func (t RSSPost) GetData() uint64     { return t.Data }

func GetRSS(setting core.SettingYaml) ([]RSSPost, error) {
	rssparse := gofeed.NewParser()
	feed, err := rssparse.ParseURL(setting.Rss.Url)
	if err != nil {
		return nil, err
	}
	var posts []RSSPost
	for _, item := range feed.Items {
		post := RSSPost{
			author:  item.Author.Name,
			content: item.Content,
			url:     item.Link,
			images:  []string{item.Image.URL},
			Data:    uint64(item.PublishedParsed.Unix()),
		}
		posts = append(posts, post)
	}
	return posts, nil

}
