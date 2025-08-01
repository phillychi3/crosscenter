package main

import (
	"crosscenter/core"
	"crosscenter/sites"
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"

	_ "github.com/joho/godotenv/autoload"
	"github.com/k0kubun/pp/v3"
	"github.com/peterbourgon/diskv/v3"
	"github.com/robfig/cron"
	"go.uber.org/zap"
)

func postToSocialMedia(poster sites.SocialMediaPoster, post sites.PostInterface, setting core.SettingYaml, db *diskv.Diskv) (string, error) {
	return poster.Post(post, setting, db)
}

// 解析路由字符串 (例如: "twitter:0" -> media="twitter", index=0)
func parseRoute(route string) (string, int) {
	parts := strings.Split(route, ":")
	if len(parts) != 2 {
		return "", -1
	}

	index, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", -1
	}

	return parts[0], index
}

// 根據路由配置獲取目標列表
func getTargetsForSource(setting core.SettingYaml, sourceRoute string) []string {
	var targets []string

	for _, route := range setting.Routing.SyncRoutes {
		if route.Enable && route.From == sourceRoute {
			targets = append(targets, route.To...)
		}
	}

	return targets
}

// 檢查特定帳號是否需要同步 (基於路由配置)
func shouldSyncAccount(setting core.SettingYaml, media string, index int) bool {
	sourceRoute := fmt.Sprintf("%s:%d", media, index)
	targets := getTargetsForSource(setting, sourceRoute)
	return len(targets) > 0
}

// 檢查特定帳號是否可以接收貼文 (基於路由配置)
func canReceivePost(setting core.SettingYaml, media string, index int) bool {
	targetRoute := fmt.Sprintf("%s:%d", media, index)

	for _, route := range setting.Routing.SyncRoutes {
		if route.Enable {
			for _, target := range route.To {
				if target == targetRoute {
					return true
				}
			}
		}
	}

	return false
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
			// 使用路由配置檢查是否需要同步此帳號
			if shouldSyncAccount(setting, strings.ToLower(media), i) {
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

		// 儲存每個來源的新貼文，按照 "media:index" 格式
		sourceNewPosts := make(map[string][]sites.PostInterface)

		// 獲取所有有路由配置的來源的新貼文
		for media, getPosts := range sites.Medias {
			getPostsFunc := getPosts.(func(core.SettingYaml) ([]sites.PostInterface, error))

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
				sourceRoute := fmt.Sprintf("%s:%d", strings.ToLower(media), i)
				if shouldSyncAccount(setting, strings.ToLower(media), i) {
					if posts, err := processMediaAccount(media, i, setting, db, getPostsFunc); err == nil && len(posts) > 0 {
						sourceNewPosts[sourceRoute] = posts
						core.Info(fmt.Sprintf("Found %d new posts from %s", len(posts), sourceRoute))
					}
				}
			}
		}

		// 根據路由配置分發貼文
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

		for sourceRoute, posts := range sourceNewPosts {
			targets := getTargetsForSource(setting, sourceRoute)

			for _, post := range posts {
				for _, targetRoute := range targets {
					targetMedia, targetIndex := parseRoute(targetRoute)
					if targetMedia == "" || targetIndex == -1 {
						core.Error(fmt.Sprintf("Invalid target route: %s", targetRoute), zap.String("route", targetRoute))
						continue
					}

					var targetKey string
					switch targetMedia {
					case "twitter":
						targetKey = "Twitter"
					case "threads":
						targetKey = "Threads"
					case "discord":
						targetKey = "Discord"
					case "bluesky":
						targetKey = "BlueSky"
					default:
						core.Error(fmt.Sprintf("Unsupported target media: %s", targetMedia), zap.String("media", targetMedia))
						continue
					}

					if site, exists := sites.PostMedias[targetKey]; exists {
						pp.Printf("Posting from %s to %s\n", sourceRoute, targetRoute)
						pp.Println(post)

						id, err := postToSocialMedia(site, post, setting, db)
						if err != nil {
							core.Error(fmt.Sprintf("Error posting from %s to %s", sourceRoute, targetRoute), zap.Error(err))
							continue
						}

						Allposts[targetKey] = append(Allposts[targetKey], id)
						core.Info(fmt.Sprintf("Successfully posted from %s to %s, id: %s", sourceRoute, targetRoute, id))
					}
				}
			}
		}

		for sitename, posts := range Allposts {
			if Bposts, err := json.Marshal(posts); err != nil {
				core.Fatal("Error marshalling media post", zap.Error(err))
			} else if err := db.Write(sitename, Bposts); err != nil {
				core.Fatal("Error writing media post", zap.Error(err))
			}
		}
	})
	c.Start()
	select {}

}
