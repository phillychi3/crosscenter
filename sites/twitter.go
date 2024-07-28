package sites

import "fmt"

// 類型 twitter

type twitteruser struct {
	username string
	token    string
}

func GetTwitterPosts(twitter twitteruser) {
	fmt.Println(twitter.username)
	fmt.Println(twitter.token)
}

func PostTwitterPost() {

}
