package main

import (
	"crosscenter/core"
	"crosscenter/sites"
	"encoding/json"
	"fmt"
	"reflect"
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

func checkSettingAtIndex(setting core.SettingYaml, mediaType string, index int, fieldName string) bool {
	settingValue := reflect.ValueOf(setting)
	var fieldValue reflect.Value

	switch mediaType {
	case "Twitter":
		fieldValue = settingValue.FieldByName("Twitter")
	case "Threads":
		fieldValue = settingValue.FieldByName("Threads")
	case "BlueSky":
		fieldValue = settingValue.FieldByName("BlueSky")
	case "Rss":
		fieldValue = settingValue.FieldByName("Rss")
	case "Discord":
		fieldValue = settingValue.FieldByName("DiscordWebhook")
	default:
		return false
	}

	if !fieldValue.IsValid() || index >= fieldValue.Len() {
		return false
	}

	item := fieldValue.Index(index)
	if field := item.FieldByName(fieldName); field.IsValid() {
		return field.Bool()
	}

	return false
}

func hasAnySyncEnabled(setting core.SettingYaml, mediaType string) bool {
	settingValue := reflect.ValueOf(setting)
	var fieldValue reflect.Value

	switch mediaType {
	case "Twitter":
		fieldValue = settingValue.FieldByName("Twitter")
	case "Threads":
		fieldValue = settingValue.FieldByName("Threads")
	case "BlueSky":
		fieldValue = settingValue.FieldByName("BlueSky")
	case "Rss":
		fieldValue = settingValue.FieldByName("Rss")
	default:
		return false
	}

	if !fieldValue.IsValid() {
		return false
	}

	for i := 0; i < fieldValue.Len(); i++ {
		item := fieldValue.Index(i)
		if enableSyncField := item.FieldByName("ENABLESYNC"); enableSyncField.IsValid() && enableSyncField.Bool() {
			return true
		}
	}

	return false
}

func hasAnyPostEnabled(setting core.SettingYaml, mediaType string) bool {
	settingValue := reflect.ValueOf(setting)
	var fieldValue reflect.Value

	switch mediaType {
	case "Twitter":
		fieldValue = settingValue.FieldByName("Twitter")
	case "Threads":
		fieldValue = settingValue.FieldByName("Threads")
	case "BlueSky":
		fieldValue = settingValue.FieldByName("BlueSky")
	case "Discord":
		fieldValue = settingValue.FieldByName("DiscordWebhook")
	default:
		return false
	}

	if !fieldValue.IsValid() {
		return false
	}

	for i := 0; i < fieldValue.Len(); i++ {
		item := fieldValue.Index(i)
		if enablePostField := item.FieldByName("ENABLEPOST"); enablePostField.IsValid() && enablePostField.Bool() {
			return true
		}
	}

	return false
}

func checkSyncSetting(setting core.SettingYaml) func(string, int) bool {
	return func(site string, i int) bool {
		return checkSettingAtIndex(setting, site, i, "ENABLESYNC")
	}
}

func checkPostSetting(setting core.SettingYaml) func(string, int) bool {
	return func(site string, i int) bool {
		return checkSettingAtIndex(setting, site, i, "ENABLEPOST")
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
		settingValue := reflect.ValueOf(setting)
		var fieldValue reflect.Value

		switch media {
		case "Twitter":
			fieldValue = settingValue.FieldByName("Twitter")
		case "Threads":
			fieldValue = settingValue.FieldByName("Threads")
		case "BlueSky":
			fieldValue = settingValue.FieldByName("BlueSky")
		case "Rss":
			fieldValue = settingValue.FieldByName("Rss")
		default:
			continue
		}

		if !fieldValue.IsValid() {
			continue
		}

		for i := 0; i < fieldValue.Len(); i++ {
			item := fieldValue.Index(i)
			if enableSyncField := item.FieldByName("ENABLESYNC"); enableSyncField.IsValid() && enableSyncField.Bool() {
				if err := initMediaAccount(media, i, setting, db, getPosts); err != nil {
					core.Error(fmt.Sprintf("Failed to initialize %s account %d", media, i), zap.Error(err))
				}
			}
		}
	}
}

func processMediaAccount(media string, index int, setting core.SettingYaml, db *diskv.Diskv, getPostsFunc func(core.SettingYaml) ([]sites.PostInterface, error)) ([]sites.PostInterface, error) {
	dbKey := fmt.Sprintf("%s_%d", media, index)

	post_history, err := db.Read(dbKey)
	if err != nil {
		post_history = []byte("[]")
	}

	var postHistory []string
	err = json.Unmarshal(post_history, &postHistory)
	if err != nil {
		core.Fatal("Error unmarshalling post history", zap.Error(err))
		return nil, err
	}

	posts, err := getPostsFunc(setting)
	if err != nil {
		core.Error(fmt.Sprintf("Error getting posts from %s_%d", media, index), zap.Error(err))
		return nil, err
	}

	var newPosts []sites.PostInterface
	for _, post := range posts {
		if !slices.Contains(postHistory, post.GetID()) {
			postHistory = append(postHistory, post.GetID())
			newPosts = append(newPosts, post)
		}
	}

	postHistoryBytes, err := json.Marshal(postHistory)
	if err != nil {
		core.Error("Error marshalling post history", zap.Error(err))
		return newPosts, err
	}

	err = db.Write(dbKey, postHistoryBytes)
	if err != nil {
		core.Error("Error writing post history to db", zap.Error(err))
	}

	return newPosts, nil
}

func getNewPostsForMedia(media string, setting core.SettingYaml, db *diskv.Diskv, getPostsFunc func(core.SettingYaml) ([]sites.PostInterface, error)) []sites.PostInterface {
	checkSync := checkSyncSetting(setting)
	var allNewPosts []sites.PostInterface

	settingValue := reflect.ValueOf(setting)
	var fieldValue reflect.Value

	switch media {
	case "Twitter":
		fieldValue = settingValue.FieldByName("Twitter")
	case "Threads":
		fieldValue = settingValue.FieldByName("Threads")
	case "BlueSky":
		fieldValue = settingValue.FieldByName("BlueSky")
	case "Rss":
		fieldValue = settingValue.FieldByName("Rss")
	default:
		return allNewPosts
	}

	if !fieldValue.IsValid() {
		return allNewPosts
	}

	for i := 0; i < fieldValue.Len(); i++ {
		if checkSync(media, i) {
			if posts, err := processMediaAccount(media, i, setting, db, getPostsFunc); err == nil {
				allNewPosts = append(allNewPosts, posts...)
			}
		}
	}

	return allNewPosts
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
			if !hasAnySyncEnabled(setting, media) {
				continue
			}
			getPostsFunc := getPosts.(func(core.SettingYaml) ([]sites.PostInterface, error))
			newPosts := getNewPostsForMedia(media, setting, db, getPostsFunc)
			if len(newPosts) > 0 {
				needsendposts[media] = newPosts
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
					if sitename == media || !hasAnyPostEnabled(setting, sitename) {
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
