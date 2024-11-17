package main

import (
	"crosscenter/core"
	"crosscenter/sites"
	"encoding/json"
	"fmt"
	"slices"

	_ "github.com/joho/godotenv/autoload"
	"github.com/k0kubun/pp/v3"
	"github.com/peterbourgon/diskv/v3"
	"github.com/robfig/cron"
	"go.uber.org/zap"
)

func postToSocialMedia(poster sites.SocialMediaPoster, post sites.PostInterface, setting core.SettingYaml, db *diskv.Diskv) (string, error) {
	return poster.Post(post, setting, db)
}

func check_get_setting(site string, setting core.SettingYaml) bool {
	switch site {
	case "Twitter":
		if !setting.Twitter.ENABLESYNC {
			return false
		} else {
			return true
		}
	case "Threads":
		if !setting.Threads.ENABLESYNC {
			return false
		} else {
			return true
		}
	case "BlueSky":
		if !setting.BlueSky.ENABLESYNC {
			return false
		} else {
			return true
		}
	case "Rss":
		if !setting.Rss.ENABLESYNC {
			return false
		} else {
			return true
		}
	default:
		return false
	}
}

func ckeck_post_setting(site string, setting core.SettingYaml) bool {
	switch site {
	case "Twitter":
		if !setting.Twitter.ENABLEPOST {
			return false
		} else {
			return true
		}
	case "Threads":
		if !setting.Threads.ENABLEPOST {
			return false
		} else {
			return true
		}
	case "Discord":
		if !setting.DiscordWebhook.ENABLEPOST {
			return false
		} else {
			return true
		}
	case "BlueSky":
		if !setting.BlueSky.ENABLEPOST {
			return false
		} else {
			return true
		}
	default:
		return false
	}
}

func _init(setting core.SettingYaml, db *diskv.Diskv) {
	for media, getPosts := range sites.Medias {
		if !check_get_setting(media, setting) {
			continue
		}
		post := []string{}
		dbpost, err := db.Read(media)
		json.Unmarshal(dbpost, &post)
		core.Info(fmt.Sprintf("db data: %d %s", len(post), media))
		if err != nil || len(post) == 0 {
			core.Info(fmt.Sprintf("first init %s", media))
			getPostsFunc := getPosts.(func(core.SettingYaml) ([]sites.PostInterface, error))
			posts, err := getPostsFunc(setting)
			if err != nil {
				core.Error(fmt.Sprintf("Error getting posts from %s\n", media), zap.Error(err))
				continue
			}
			core.Info(fmt.Sprintf("Get %d posts from %s", len(posts), media))
			postHistory := []string{}
			for _, post := range posts {
				postHistory = append(postHistory, post.GetID())
			}
			postHistoryBytes, err := json.Marshal(postHistory)
			if err != nil {
				core.Error("Error marshalling post history", zap.Error(err))
				continue
			}
			err = db.Write(media, postHistoryBytes)
			if err != nil {
				core.Fatal("Error writing post history to db", zap.Error(err))
			}
		}
	}
}

func main() {
	setting := core.LoadSetting()
	db := core.Getdb()
	_init(setting, db)
	c := cron.New()
	c.AddFunc("@hourly", func() {

		core.Info("Start get post")

		needsendposts := make(map[string][]sites.PostInterface)

		for media, getPosts := range sites.Medias {
			if !check_get_setting(media, setting) {
				continue
			}
			getPostsFunc := getPosts.(func(core.SettingYaml) ([]sites.PostInterface, error))
			{
				posts, err := getPostsFunc(setting)
				if err != nil {
					core.Error(fmt.Sprintf("Error getting posts from %s", media), zap.Error(err))
					continue
				}
				post_history, err := db.Read(media)
				if err != nil {
					post_history = []byte("[]")
				}
				var postHistory []string
				err = json.Unmarshal(post_history, &postHistory)
				if err != nil {
					core.Fatal("Error unmarshalling post history", zap.Error(err))
				}

				for _, post := range posts {
					if !slices.Contains(postHistory, post.GetID()) {
						postHistory = append(postHistory, post.GetID())
						needsendposts[media] = append(needsendposts[media], post)
					}
				}

				// 塞回去
				postHistoryBytes, err := json.Marshal(postHistory)
				if err != nil {
					core.Error("Error marshalling post history", zap.Error(err))
					continue
				}
				err = db.Write(media, postHistoryBytes)
				if err != nil {
					core.Error("Error writing post history to db", zap.Error(err))
				}
			}
		}

		Allposts := make(map[string][]string)
		for sitename := range sites.PostMedias {
			var posts []string
			if bposts, err := db.Read(sitename); err == nil {
				if err := json.Unmarshal(bposts, &posts); err != nil {
					core.Fatal("Error unmarshalling media post", zap.Error(err))
				}
			}
			Allposts[sitename] = posts
		}

		for media, posts := range needsendposts {
			for _, post := range posts {
				for sitename, site := range sites.PostMedias {
					//不需要發給自己
					if sitename == media || !ckeck_post_setting(sitename, setting) {
						continue
					}
					pp.Println(post)
					id, err := postToSocialMedia(site, post, setting, db)
					if err != nil {
						core.Error(fmt.Sprintf("Error posting to social media: %s \n", sitename), zap.Error(err))
						continue
					}
					Allposts[sitename] = append(Allposts[sitename], id)
					core.Info(fmt.Sprintf("success post to %s id: %s\n", sitename, id))
				}
			}
		}

		for sitename, posts := range Allposts {
			if Bposts, err := json.Marshal(posts); err != nil {
				core.Fatal("Error marshalling media pos", zap.Error(err))
			} else if err := db.Write(sitename, Bposts); err != nil {
				core.Fatal("Error writing media post", zap.Error(err))
			}
		}
	})
	c.Start()
	select {}

}
