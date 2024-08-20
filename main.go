package main

import (
	"crosscenter/core"
	"crosscenter/sites"

	_ "github.com/joho/godotenv/autoload"
	"github.com/k0kubun/pp/v3"
	"github.com/peterbourgon/diskv/v3"
)

func postToSocialMedia(poster sites.SocialMediaPoster, post sites.PostInterface, setting core.SettingYaml, db *diskv.Diskv) error {
	return poster.Post(post, setting, db)
}

func main() {
	setting := core.LoadSetting()
	// db := core.Getdb()

	res, err := sites.GetThreadsPosts(setting)
	if err != nil {
		panic(err)
	}
	pp.Print(res)

	// err := postToSocialMedia(&sites.TwitterPoster{}, testpost, setting, db)
	// if err != nil {
	// 	panic(err)
	// }

}
