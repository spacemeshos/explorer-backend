package cache

import (
	"github.com/patrickmn/go-cache"
	"time"
)

var Cache *cache.Cache

func Init() {
	Cache = cache.New(1*time.Hour, 10*time.Minute)
}
