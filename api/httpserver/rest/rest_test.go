package rest_test

import (
	"context"
	"fmt"
	"github.com/spacemeshos/explorer-backend/model"
	"os"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/spacemeshos/explorer-backend/internal/storage/storagereader"
	"github.com/spacemeshos/explorer-backend/storage"
	"github.com/spacemeshos/explorer-backend/test/testseed"
	"github.com/spacemeshos/explorer-backend/test/testserver"
)

const testAPIServiceDB = "explorer_test"

var (
	apiServer *testserver.TestAPIService
	generator *testseed.SeedGenerator
	seed      *testseed.TestServerSeed
	dbPort    = 27017
)

func TestMain(m *testing.M) {
	mongoURL := fmt.Sprintf("mongodb://localhost:%d", dbPort)
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURL))
	if err != nil {
		fmt.Println("failed to connect to mongo", err)
		os.Exit(1)
	}
	database := client.Database(testAPIServiceDB)
	if database != nil {
		if err = database.Drop(context.Background()); err != nil {
			fmt.Println("failed to drop db", err)
			os.Exit(1)
		}
	}

	db, err := storage.New(context.Background(), mongoURL, testAPIServiceDB)
	if err != nil {
		fmt.Println("failed to init storage to mongo", err)
		os.Exit(1)
	}
	seed = testseed.GetServerSeed()
	db.OnNetworkInfo(seed.NetID, seed.GenesisTime, seed.EpochNumLayers, seed.MaxTransactionPerSecond, seed.LayersDuration, seed.GetPostUnitsSize())

	dbReader, err := storagereader.NewStorageReader(context.Background(), mongoURL, testAPIServiceDB)
	if err != nil {
		fmt.Println("failed to init storage to mongo", err)
		os.Exit(1)
	}

	apiServer, err = testserver.StartTestAPIServiceV2(db, dbReader)
	// old version of app here apiServer, err = testserver.StartTestAPIService(dbPort, db)
	if err != nil {
		fmt.Println("failed to start test api service", err)
		os.Exit(1)
	}
	generator = testseed.NewSeedGenerator(testseed.GetServerSeed())
	if err = generator.GenerateEpoches(10); err != nil {
		fmt.Println("failed to generate epochs", err)
		os.Exit(1)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	if err = generator.SaveEpoches(ctx, db); err != nil {
		fmt.Println("failed to save generated epochs", err)
		os.Exit(1)
	}

	code := m.Run()
	db.Close()
	os.Exit(code)
}

type layerResp struct {
	Data       []model.Layer `json:"data"`
	Pagination pagination    `json:"pagination"`
}

type epochResp struct {
	Data       []model.Epoch `json:"data"`
	Pagination pagination    `json:"pagination"`
}

type transactionResp struct {
	Data       []model.Transaction `json:"data"`
	Pagination pagination          `json:"pagination"`
}

type smesherResp struct {
	Data       []model.Smesher `json:"data"`
	Pagination pagination      `json:"pagination"`
}

type rewardResp struct {
	Data       []model.Reward `json:"data"`
	Pagination pagination     `json:"pagination"`
}

type accountResp struct {
	Data       []model.Account `json:"data"`
	Pagination pagination      `json:"pagination"`
}

type atxResp struct {
	Data       []model.Activation `json:"data"`
	Pagination pagination         `json:"pagination"`
}
type blockResp struct {
	Data       []model.Block `json:"data"`
	Pagination pagination    `json:"pagination"`
}

type appResp struct {
	Data       []model.App `json:"data"`
	Pagination pagination  `json:"pagination"`
}

type redirect struct {
	Redirect string `json:"redirect"`
}

type pagination struct {
	TotalCount  int  `json:"totalCount"`
	PageCount   int  `json:"pageCount"`
	PerPage     int  `json:"perPage"`
	Next        int  `json:"next"`
	HasNext     bool `json:"hasNext"`
	Current     int  `json:"current"`
	Previous    int  `json:"previous"`
	HasPrevious bool `json:"hasPrevious"`
}
