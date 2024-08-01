package main

import (
	"crosscenter/sites"
	"fmt"
)

func main() {
	threadeuser := sites.Threadsuser{
		Username: "instagram",
	}
	result, err := sites.GetThreadsPosts(threadeuser)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(result)
}
