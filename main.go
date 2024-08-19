package main

import (
	"crosscenter/core"
	"crosscenter/sites"
	"fmt"

	_ "github.com/joho/godotenv/autoload"
	"github.com/k0kubun/pp/v3"
)

func main() {
	// threadeuser := sites.Threadsuser{
	// 	Username: "instagram",
	// }
	// result, err := sites.GetThreadsPosts(threadeuser)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// fmt.Println(result)
	setting := core.LoadSetting()
	result, err := sites.GetTwitterPosts(setting)
	if err != nil {
		fmt.Println(err)
	}
	pp.Print(result)

}
