package collector

import (
	"fmt"
	"github.com/labstack/echo/v4"
	pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
	"github.com/spacemeshos/explorer-backend/model"
	"github.com/spacemeshos/go-spacemesh/common/types"
	"github.com/spacemeshos/go-spacemesh/log"
	"net/http"
	"strconv"
)

func (c *Collector) StartHttpServer(apiHost string, apiPort int) {
	e := echo.New()

	e.GET("/sync/atx/:id", func(ctx echo.Context) error {
		id := ctx.Param("id")

		log.Info("http syncing atx %s", id)
		go func() {
			atx, err := c.dbClient.GetAtxById(c.db, id)
			if err != nil {
				log.Warning("syncing atx %s failed with error %d", id, err)
				return
			}
			if atx != nil {
				c.listener.OnActivation(atx)
				c.listener.RecalculateEpochStats()
			}
		}()

		return ctx.NoContent(http.StatusOK)
	})

	e.GET("/sync/atxs/ts/:ts", func(ctx echo.Context) error {
		ts := ctx.Param("ts")
		timestamp, err := strconv.ParseInt(ts, 10, 64)
		if err != nil {
			return ctx.String(http.StatusBadRequest, "Invalid parameter")
		}

		log.Info("http syncing atxs from %d", timestamp)
		go func() {
			err = c.dbClient.GetAtxsReceivedAfter(c.db, timestamp, func(atx *types.VerifiedActivationTx) bool {
				c.listener.OnActivation(atx)
				return true
			})
			if err != nil {
				log.Warning("syncing atxs from %s failed with error %d", ts, err)
				return
			}
			c.listener.RecalculateEpochStats()
		}()

		return ctx.NoContent(http.StatusOK)
	})

	e.GET("/sync/bulk/atxs/ts/:ts", func(ctx echo.Context) error {
		ts := ctx.Param("ts")
		timestamp, err := strconv.ParseInt(ts, 10, 64)
		if err != nil {
			return ctx.String(http.StatusBadRequest, "Invalid parameter")
		}

		log.Info("http syncing atxs from %d", timestamp)
		go func() {
			var atxs []*model.Activation
			err = c.dbClient.GetAtxsReceivedAfter(c.db, timestamp, func(atx *types.VerifiedActivationTx) bool {
				atxs = append(atxs, model.NewActivation(atx))
				return true
			})
			if err != nil {
				log.Warning("syncing atxs from %s failed with error %d", ts, err)
				return
			}
			c.listener.OnActivations(atxs)
			c.listener.RecalculateEpochStats()
		}()

		return ctx.NoContent(http.StatusOK)
	})

	e.GET("/sync/atxs/:epoch", func(ctx echo.Context) error {
		epoch := ctx.Param("epoch")
		epochId, err := strconv.ParseInt(epoch, 10, 64)
		if err != nil {
			return ctx.String(http.StatusBadRequest, "Invalid parameter")
		}

		log.Info("http syncing atxs for epoch %s", epoch)
		go func() {
			err = c.dbClient.GetAtxsByEpoch(c.db, epochId, func(atx *types.VerifiedActivationTx) bool {
				c.listener.OnActivation(atx)
				return true
			})
			if err != nil {
				log.Warning("syncing atxs for %s failed with error %d", epoch, err)
				return
			}
			c.listener.RecalculateEpochStats()
		}()

		return ctx.NoContent(http.StatusOK)
	})

	e.GET("/sync/bulk/atxs/:epoch", func(ctx echo.Context) error {
		epoch := ctx.Param("epoch")
		epochId, err := strconv.ParseInt(epoch, 10, 64)
		if err != nil {
			return ctx.String(http.StatusBadRequest, "Invalid parameter")
		}

		log.Info("http syncing atxs for epoch %s", epoch)
		go func() {
			count, err := c.dbClient.CountAtxsByEpoch(c.db, epochId)
			if err != nil {
				log.Warning("syncing atxs for %s failed with error %d", epoch, err)
				return
			}
			batchSize := 100000
			totalPages := (count + batchSize - 1) / batchSize
			for page := 0; page < totalPages; page++ {
				offset := page * batchSize
				var atxs []*model.Activation
				err = c.dbClient.GetAtxsByEpochPaginated(c.db, epochId, int64(batchSize), int64(offset), func(atx *types.VerifiedActivationTx) bool {
					atxs = append(atxs, model.NewActivation(atx))
					return true
				})
				if err != nil {
					log.Warning("syncing atxs for %s failed with error %d", epoch, err)
					return
				}
				c.listener.OnActivations(atxs)
				atxs = nil
			}
			c.listener.RecalculateEpochStats()
		}()

		return ctx.NoContent(http.StatusOK)
	})

	e.GET("/sync/layer/:layer", func(ctx echo.Context) error {
		layer := ctx.Param("layer")
		layerId, err := strconv.ParseInt(layer, 10, 64)
		if err != nil {
			return ctx.String(http.StatusBadRequest, "Invalid parameter")
		}
		lid := types.LayerID(layerId)

		go func() {
			l, err := c.dbClient.GetLayer(c.db, lid, c.listener.GetEpochNumLayers())
			if err != nil {
				log.Warning("%v", err)
				return
			}

			log.Info("http syncing layer: %d", l.Number.Number)
			c.listener.OnLayer(l)
		}()

		return ctx.NoContent(http.StatusOK)
	})

	e.GET("/sync/rewards/:layer", func(ctx echo.Context) error {
		layer := ctx.Param("layer")
		layerId, err := strconv.ParseInt(layer, 10, 64)
		if err != nil {
			return ctx.String(http.StatusBadRequest, "Invalid parameter")
		}
		lid := types.LayerID(layerId)

		go func() {
			log.Info("http syncing rewards for layer: %d", lid.Uint32())
			rewards, err := c.dbClient.GetLayerRewards(c.db, lid)
			if err != nil {
				log.Warning("%v", err)
				return
			}

			for _, reward := range rewards {
				r := &pb.Reward{
					Layer:       &pb.LayerNumber{Number: reward.Layer.Uint32()},
					Total:       &pb.Amount{Value: reward.TotalReward},
					LayerReward: &pb.Amount{Value: reward.LayerReward},
					Coinbase:    &pb.AccountId{Address: reward.Coinbase.String()},
					Smesher:     &pb.SmesherId{Id: reward.SmesherID.Bytes()},
				}
				c.listener.OnReward(r)
			}

			c.listener.UpdateEpochStats(lid.Uint32())
		}()

		return ctx.NoContent(http.StatusOK)
	})

	e.Logger.Fatal(e.Start(fmt.Sprintf("%s:%d", apiHost, apiPort)))
}
