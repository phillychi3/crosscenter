package sites

import (
	"crosscenter/core"
)

type BSKYPost struct {
	author  string
	content string
	url     string
	images  []string
	Data    uint64
	Id      string
}

func (t BSKYPost) GetAuthor() string   { return t.author }
func (t BSKYPost) GetContent() string  { return t.content }
func (t BSKYPost) GetURL() string      { return t.url }
func (t BSKYPost) GetImages() []string { return t.images }
func (t BSKYPost) GetDate() uint64     { return t.Data }
func (t BSKYPost) GetID() string       { return t.Id }

func GetBSKY(setting core.SettingYaml) ([]PostInterface, error) {

	var posts []PostInterface
	return posts, nil

}
