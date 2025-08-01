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

func check_get_setting(site string, i int, setting core.SettingYaml) bool {
	switch site {
	case "Twitter":
		if i >= len(setting.Twitter) {
			return false
		}
		return setting.Twitter[i].ENABLESYNC
	case "Threads":
		if i >= len(setting.Threads) {
			return false
		}
		return setting.Threads[i].ENABLESYNC
	case "BlueSky":
		if i >= len(setting.BlueSky) {
			return false
		}
		return setting.BlueSky[i].ENABLESYNC
	case "Rss":
		if i >= len(setting.Rss) {
			return false
		}
		return setting.Rss[i].ENABLESYNC
	default:
		return false
	}
}

func ckeck_post_setting(site string, i int, setting core.SettingYaml) bool {
	switch site {
	case "Twitter":
		if i >= len(setting.Twitter) {
			return false
		}
		return setting.Twitter[i].ENABLEPOST
	case "Threads":
		if i >= len(setting.Threads) {
			return false
		}
		return setting.Threads[i].ENABLEPOST
	case "Discord":
		if i >= len(setting.DiscordWebhook) {
			return false
		}
		return setting.DiscordWebhook[i].ENABLEPOST
	case "BlueSky":
		if i >= len(setting.BlueSky) {
			return false
		}
		return setting.BlueSky[i].ENABLEPOST
	default:
		return false
	}
}

func initMediaAccount(media string, i int, setting core.SettingYaml, db *diskv.Diskv, getPosts interface{}) error {
	dbKey := fmt.Sprintf("%s_%d", media, i)
	post := []string{}
	dbpost, err := db.Read(dbKey)
	json.Unmarshal(dbpost, &post)
	core.Info(fmt.Sprintf("db data: %d %s_%d", len(post), media, i))

	if err != nil || len(post) == 0 {
		core.Info(fmt.Sprintf("first init %s_%d", media, i))
		getPostsFunc := getPosts.(func(core.SettingYaml) ([]sites.PostInterface, error))
		posts, err := getPostsFunc(setting)
		if err != nil {
			core.Error(fmt.Sprintf("Error getting posts from %s_%d\n", media, i), zap.Error(err))
			return err
		}

		core.Info(fmt.Sprintf("Get %d posts from %s_%d", len(posts), media, i))
		postHistory := []string{}
		for _, post := range posts {
			postHistory = append(postHistory, post.GetID())
		}

		postHistoryBytes, err := json.Marshal(postHistory)
		if err != nil {
			core.Error("Error marshalling post history", zap.Error(err))
			return err
		}

		err = db.Write(dbKey, postHistoryBytes)
		if err != nil {
			core.Fatal("Error writing post history to db", zap.Error(err))
			return err
		}
	}
	return nil
}

func _init(setting core.SettingYaml, db *diskv.Diskv) {
	for media, getPosts := range sites.Medias {
		switch media {
		case "Twitter":
			for i, twitterSetting := range setting.Twitter {
				if !twitterSetting.ENABLESYNC {
					continue
				}
				if err := initMediaAccount(media, i, setting, db, getPosts); err != nil {
					core.Error(fmt.Sprintf("Failed to initialize %s account %d", media, i), zap.Error(err))
				}
			}
		case "Threads":
			for i, threadsSetting := range setting.Threads {
				if !threadsSetting.ENABLESYNC {
					continue
				}
				if err := initMediaAccount(media, i, setting, db, getPosts); err != nil {
					core.Error(fmt.Sprintf("Failed to initialize %s account %d", media, i), zap.Error(err))
				}
			}
		case "BlueSky":
			for i, blueSkySetting := range setting.BlueSky {
				if !blueSkySetting.ENABLESYNC {
					continue
				}
				if err := initMediaAccount(media, i, setting, db, getPosts); err != nil {
					core.Error(fmt.Sprintf("Failed to initialize %s account %d", media, i), zap.Error(err))
				}
			}
		case "Rss":
			for i, rssSetting := range setting.Rss {
				if !rssSetting.ENABLESYNC {
					continue
				}
				if err := initMediaAccount(media, i, setting, db, getPosts); err != nil {
					core.Error(fmt.Sprintf("Failed to initialize %s account %d", media, i), zap.Error(err))
				}
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
