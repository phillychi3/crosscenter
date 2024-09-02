package sites

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"crosscenter/core"

	"github.com/peterbourgon/diskv/v3"
)

type DiscordPoster struct{}

func (dp DiscordPoster) Post(post PostInterface, setting core.SettingYaml, db *diskv.Diskv) (string, error) {
	return PostDiscordWebhook(post, setting)
}

type discordthumbnail struct {
	Url string `json:"url"`
}
type discordfooter struct {
	Text     string `json:"text"`
	Icon_url string `json:"icon_url"`
}

type discordembed struct {
	Title       string           `json:"title"`
	Url         string           `json:"url"`
	Description string           `json:"description"`
	Color       int              `json:"color"`
	Thumbnail   discordthumbnail `json:"thumbnail"`
	Footer      discordfooter    `json:"footer"`
	Timestamp   time.Time        `json:"timestamp"`
}

type sendwebhook struct {
	Username string         `json:"username"`
	Avatar   string         `json:"avatar_url"`
	Embeds   []discordembed `json:"embeds"`
}

func PostDiscordWebhook(post PostInterface, setting core.SettingYaml) (string, error) {
	core.Debug("Posting to Discord Webhook")
	url := setting.DiscordWebhook.Url
	var embed discordembed
	if len(post.GetImages()) != 0 {
		embed = discordembed{
			Title:       post.GetAuthor(),
			Url:         post.GetURL(),
			Description: post.GetContent(),
			Color:       0x00ff00,
			Footer: discordfooter{
				Text: setting.DiscordWebhook.FooterText,
			},
			Thumbnail: discordthumbnail{
				Url: post.GetImages()[0],
			},
			Timestamp: time.Unix(int64(post.GetDate()), 0),
		}
	} else {
		embed = discordembed{
			Title:       post.GetAuthor(),
			Url:         post.GetURL(),
			Description: post.GetContent(),
			Color:       0x00ff00,
			Footer: discordfooter{
				Text: setting.DiscordWebhook.FooterText,
			},
			Timestamp: time.Unix(int64(post.GetDate()), 0),
		}
	}
	var username string
	if setting.DiscordWebhook.Username != "" {
		username = setting.DiscordWebhook.Username
	} else {
		username = post.GetAuthor()
	}

	sendhook := sendwebhook{
		Username: username,
		Avatar:   setting.DiscordWebhook.AvatarUrl,
		Embeds:   []discordembed{embed},
	}

	payload, err := json.Marshal(sendhook)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 429 {
		core.Error("Rate limit exceeded, waiting 5 seconds")
		time.Sleep(5 * time.Second)
		return PostDiscordWebhook(post, setting)
	}
	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		return "", err
	}

	return "discordwebhook", nil

}
