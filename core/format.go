package core

import (
	"regexp"
	"time"
)

type PostInterface interface {
	GetAuthor() string
	GetContent() string
	GetURL() string
	GetImages() []string
	GetDate() uint64
	GetID() string
}

func Removet_co(text string) string {
	re := regexp.MustCompile(`https://t.co/[a-zA-Z0-9]+`)
	return re.ReplaceAllString(text, "")
}

func formatString(template string, data map[string]string) string {
	re := regexp.MustCompile(`\{([^}]+)\}`)
	result := re.ReplaceAllStringFunc(template, func(match string) string {
		key := match[1 : len(match)-1]
		if value, exists := data[key]; exists {
			return value
		}
		return match
	})
	return result
}

// Example: "{text} #sometag author: {author} url:{url} Date:{date}"
func TextFormat(text string, post PostInterface) string {

	data := map[string]string{
		"author": post.GetAuthor(),
		"text":   post.GetContent(),
		"url":    post.GetURL(),
		"date":   time.Unix(int64(post.GetDate()), 0).Format("2006-01-02 15:04:05"),
	}
	return formatString(text, data)
}
