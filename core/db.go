package core

import (
	"github.com/peterbourgon/diskv/v3"
)

func db() *diskv.Diskv {
	d := diskv.New(diskv.Options{
		BasePath:     "dbdata",
		CacheSizeMax: 1024 * 1024,
	})
	return d
}
