package cache

import (
	"github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/lib/v4/marshaler"
	gocacheStore "github.com/eko/gocache/store/go_cache/v4"
	gocache "github.com/patrickmn/go-cache"
	"time"
)

func New() *marshaler.Marshaler {
	client := gocache.New(gocache.NoExpiration, 6*time.Hour)
	s := gocacheStore.NewGoCache(client)
	manager := cache.New[any](s)
	return marshaler.New(manager)
}
