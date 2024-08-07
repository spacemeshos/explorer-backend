package cache

import (
	"github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/lib/v4/marshaler"
	gocacheStore "github.com/eko/gocache/store/go_cache/v4"
	gocache "github.com/patrickmn/go-cache"
	"time"
)

var Expiration time.Duration = 0
var ShortExpiration = 5 * time.Minute

func New() *marshaler.Marshaler {
	client := gocache.New(Expiration, 6*time.Hour)
	s := gocacheStore.NewGoCache(client)
	manager := cache.New[any](s)
	return marshaler.New(manager)
}
