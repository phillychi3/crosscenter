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
	GetData() uint64
	GetID() string
}

type SocialMediaPoster interface {
	Post(post PostInterface, setting core.SettingYaml, db *diskv.Diskv) error
}
