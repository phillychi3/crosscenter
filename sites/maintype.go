package sites

import (
	"crosscenter/core"
	"fmt"

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

type PostFunc func(post PostInterface, setting core.SettingYaml, db *diskv.Diskv) (string, error)

type SocialMediaPoster interface {
	Post(post PostInterface, setting core.SettingYaml, db *diskv.Diskv) (string, error)
}

var Medias = map[string]interface{}{
	"Twitter": GetTwitterPosts,
	"Threads": GetThreadsPosts,
	"Rss":     GetRSS,
	"BlueSky": GetBSKY,
}

func createTwitterAdapter() PostFunc {
	return func(post PostInterface, setting core.SettingYaml, db *diskv.Diskv) (string, error) {
		if len(setting.Twitter) == 0 {
			return "", fmt.Errorf("no Twitter settings found")
		}
		poster := TwitterPoster{}
		return poster.Post(post, setting.Twitter[0], db)
	}
}

func createThreadsAdapter() PostFunc {
	return func(post PostInterface, setting core.SettingYaml, db *diskv.Diskv) (string, error) {
		if len(setting.Threads) == 0 {
			return "", fmt.Errorf("no Threads settings found")
		}
		poster := ThreadsPoster{}
		return poster.Post(post, setting.Threads[0], db)
	}
}

func createDiscordAdapter() PostFunc {
	return func(post PostInterface, setting core.SettingYaml, db *diskv.Diskv) (string, error) {
		if len(setting.DiscordWebhook) == 0 {
			return "", fmt.Errorf("no Discord settings found")
		}
		poster := DiscordPoster{}
		return poster.Post(post, setting.DiscordWebhook[0], db)
	}
}

func createBlueSkyAdapter() PostFunc {
	return func(post PostInterface, setting core.SettingYaml, db *diskv.Diskv) (string, error) {
		if len(setting.BlueSky) == 0 {
			return "", fmt.Errorf("no BlueSky settings found")
		}
		poster := BlueSkyPoster{}
		return poster.Post(post, setting.BlueSky[0], db)
	}
}

type adapterWrapper struct {
	postFunc PostFunc
}

func (a adapterWrapper) Post(post PostInterface, setting core.SettingYaml, db *diskv.Diskv) (string, error) {
	return a.postFunc(post, setting, db)
}

var PostMedias = map[string]SocialMediaPoster{
	"Twitter": adapterWrapper{createTwitterAdapter()},
	"Threads": adapterWrapper{createThreadsAdapter()},
	"Discord": adapterWrapper{createDiscordAdapter()},
	"BlueSky": adapterWrapper{createBlueSkyAdapter()},
}
