package sites

import (
	"crosscenter/core"

	"github.com/peterbourgon/diskv/v3"
)

type PostInterface interface {
	GetAuthor() string
	GetContent() string
	GetURL() string
	GetImages() []string
	GetDate() uint64
	GetID() string
}

type SocialMediaPoster interface {
	Post(post PostInterface, setting core.SettingYaml, db *diskv.Diskv) (string, error)
}

var Medias = map[string]interface{}{
	"Twitter": GetTwitterPosts,
	"Threads": GetThreadsPosts,
	"Rss":     GetRSS,
}

var PostMedias = map[string]SocialMediaPoster{
	"Twitter": TwitterPoster{},
	"Threads": ThreadsPoster{},
}
