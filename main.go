package main

import (
	"crosscenter/core"
	"crosscenter/sites"
	"encoding/json"
	"fmt"
	"slices"

	_ "github.com/joho/godotenv/autoload"
	"github.com/peterbourgon/diskv/v3"
	"github.com/robfig/cron"
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
		fmt.Println("post count:", len(post), media)
		if err != nil || len(post) == 0 {
			fmt.Println("first init " + media)
			getPostsFunc := getPosts.(func(core.SettingYaml) ([]sites.PostInterface, error))
			posts, err := getPostsFunc(setting)
			if err != nil {
				fmt.Printf("Error getting posts from %s: %s\n", media, err)
				continue
			}
			fmt.Printf("Get %d posts from %s\n", len(posts), media)
			postHistory := []string{}
			for _, post := range posts {
				postHistory = append(postHistory, post.GetID())
			}
			postHistoryBytes, err := json.Marshal(postHistory)
			if err != nil {
				fmt.Println("Error marshalling post history:", err)
				continue
			}
			err = db.Write(media, postHistoryBytes)
			if err != nil {
				fmt.Println("Error writing post history to db:", err)
			}
		}
	}
}

func main() {
	setting := core.LoadSetting()
	db := core.Getdb()
	c := cron.New()
	_init(setting, db)
	c.AddFunc("@hourly", func() {

		fmt.Println("Start get post")

		needsendposts := make(map[string][]sites.PostInterface)

		for media, getPosts := range sites.Medias {
			if !check_get_setting(media, setting) {
				continue
			}
			getPostsFunc := getPosts.(func(core.SettingYaml) ([]sites.PostInterface, error))
			{
				posts, err := getPostsFunc(setting)
				if err != nil {
					fmt.Println("Error getting posts:", err)
					continue
				}
				post_history, err := db.Read(media)
				if err != nil {
					post_history = []byte("[]")
				}
				var postHistory []string
				err = json.Unmarshal(post_history, &postHistory)
				if err != nil {
					fmt.Println("Error on get posy history")
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
					fmt.Println("Error marshalling post history:", err)
					continue
				}
				err = db.Write(media, postHistoryBytes)
				if err != nil {
					fmt.Println("Error writing post history to db:", err)
				}
			}
		}

		for media, posts := range needsendposts {
			Bmediaposts, err := db.Read(media)
			mediaposts := []string{}
			if err != nil {
				fmt.Println("Error reading media post:", err)
			}
			err = json.Unmarshal(Bmediaposts, &mediaposts)
			if err != nil {
				fmt.Println("Error unmarshalling media post:", err)
			}
			for _, post := range posts {
				for sitename, site := range sites.PostMedias {
					//不需要發給自己
					if sitename == media || !ckeck_post_setting(sitename, setting) {
						continue
					}
					id, err := postToSocialMedia(site, post, setting, db)
					if err != nil {
						fmt.Println("Error posting to social media:", sitename, err)
						continue
					}
					mediaposts = append(mediaposts, id)
					fmt.Printf("success post to %s id: %s\n", sitename, id)
				}
			}
			Bmediaposts, err = json.Marshal(mediaposts)
			if err != nil {
				fmt.Println("Error marshalling media post:", err)
			}
			err = db.Write(media, Bmediaposts)
			if err != nil {
				fmt.Println("Error writing media post:", err)
			}
		}
	})
	c.Start()
	select {}

}
