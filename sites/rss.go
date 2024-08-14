package sites

import "github.com/mmcdole/gofeed"

type RSS struct {
	Sitename string
	Url      string
}

type RSSPost struct {
	author  string
	context string
	url     string
	images  []string
	Data    uint64
}

func (t RSSPost) GetAuthor() string   { return t.author }
func (t RSSPost) GetContext() string  { return t.context }
func (t RSSPost) GetURL() string      { return t.url }
func (t RSSPost) GetImages() []string { return t.images }
func (t RSSPost) GetData() uint64     { return t.Data }

func GetRSS(rss RSS) ([]RSSPost, error) {
	rssparse := gofeed.NewParser()
	feed, err := rssparse.ParseURL(rss.Url)
	if err != nil {
		return nil, err
	}
	var posts []RSSPost
	for _, item := range feed.Items {
		post := RSSPost{
			author:  item.Author.Name,
			context: item.Content,
			url:     item.Link,
			images:  []string{item.Image.URL},
			Data:    uint64(item.PublishedParsed.Unix()),
		}
		posts = append(posts, post)
	}
	return posts, nil

}
