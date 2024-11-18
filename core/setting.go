package core

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type SettingYaml struct {
	Twitter struct {
		Username          string `yaml:"Username"`
		PostText          string `yaml:"PostText"`
		ENABLESYNC        bool   `yaml:"ENABLE_SYNC"`
		ENABLEPOST        bool   `yaml:"ENABLE_POST"`
		CONSUMERKEY       string `yaml:"CONSUMER_KEY"`
		CONSUMERSECRET    string `yaml:"CONSUMER_SECRET"`
		ACCESSTOKEN       string `yaml:"ACCESS_TOKEN"`
		ACCESSTOKENSECRET string `yaml:"ACCESS_TOKEN_SECRET"`
		Auth_token        string `yaml:"Auth_token"`
		Ct0               string `yaml:"Ct0"`
	} `yaml:"twitter"`
	Threads struct {
		Username     string `yaml:"Username"`
		PostText     string `yaml:"PostText"`
		ENABLESYNC   bool   `yaml:"ENABLE_SYNC"`
		ENABLEPOST   bool   `yaml:"ENABLE_POST"`
		ClientSecret string `yaml:"Client_Secret"`
		AccessToken  string `yaml:"Access_Token"`
	} `yaml:"threads"`
	Rss struct {
		Sitename   string `yaml:"Sitename"`
		Url        string `yaml:"Url"`
		ENABLESYNC bool   `yaml:"ENABLE_SYNC"`
	} `yaml:"rss"`
	DiscordWebhook struct {
		Url        string `yaml:"Url"`
		AvatarUrl  string `yaml:"Avatar_Url"`
		Username   string `yaml:"Username"`
		FooterText string `yaml:"FooterText"`
		ENABLEPOST bool   `yaml:"ENABLE_POST"`
	} `yaml:"discord_webhook"`
	BlueSky struct {
		DID        string `yaml:"DID"`
		Password   string `yaml:"Password"`
		PostText   string `yaml:"PostText"`
		ENABLESYNC bool   `yaml:"ENABLE_SYNC"`
		ENABLEPOST bool   `yaml:"ENABLE_POST"`
	} `yaml:"bluesky"`
}

func LoadSetting() SettingYaml {
	Setting := SettingYaml{}
	yamlfile, err := os.ReadFile("set.yaml")
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	err = yaml.Unmarshal(yamlfile, &Setting)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	return Setting
}
