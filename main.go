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

func postToSocialMedia(poster sites.SocialMediaPoster, post sites.PostInterface, setting core.SettingYaml, db *diskv.Diskv) error {
	return poster.Post(post, setting, db)
}

func main() {
	setting := core.LoadSetting()
	db := core.Getdb()
	c := cron.New()

	c.AddFunc("@hourly", func() {

		needsendposts := make(map[string][]sites.PostInterface)

		for media, getPosts := range sites.Medias {
			getPostsFunc := getPosts.(func(core.SettingYaml) []sites.PostInterface)
			{
				posts := getPostsFunc(setting)
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
			for _, post := range posts {
				err := postToSocialMedia(sites.PostMedias[media], post, setting, db)
				if err != nil {
					fmt.Println("Error posting to social media:", err)
				}
			}
		}

	})
	c.Start()
	select {}

}
