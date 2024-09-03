package cache

import (
	"github.com/eko/gocache/lib/v4/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"time"

	"github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/lib/v4/marshaler"
	"github.com/eko/gocache/lib/v4/store"
	gocacheStore "github.com/eko/gocache/store/go_cache/v4"
	redis_store "github.com/eko/gocache/store/redis/v4"
	gocache "github.com/patrickmn/go-cache"
	"github.com/redis/go-redis/v9"

	"github.com/spacemeshos/go-spacemesh/log"
)

var (
	RedisAddress                  = ""
	Expiration      time.Duration = 0
	ShortExpiration               = 5 * time.Minute
	promMetrics                   = metrics.NewPrometheus("explorer_cache")
	LastUpdated                   = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "explorer_cache_last_updated",
			Help: "The last time the cache was updated, labeled by endpoint",
		},
		[]string{"endpoint"},
	)
)

func New() *marshaler.Marshaler {
	prometheus.MustRegister(LastUpdated)
	var manager *cache.MetricCache[any]
	if RedisAddress != "" {
		log.Info("using redis cache")
		redisStore := redis_store.NewRedis(redis.NewClient(&redis.Options{
			Addr: RedisAddress,
		}), store.WithExpiration(Expiration))
		manager = cache.NewMetric[any](
			promMetrics,
			cache.New[any](redisStore),
		)
	} else {
		log.Info("using memory cahe")
		client := gocache.New(Expiration, 6*time.Hour)
		s := gocacheStore.NewGoCache(client)
		manager = cache.NewMetric[any](
			promMetrics,
			cache.New[any](s),
		)
	}

	return marshaler.New(manager)
}
