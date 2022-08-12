package tests

import (
	"context"
	"fmt"
	"os"
	"testing"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/spacemeshos/explorer-backend/storage"
	"github.com/spacemeshos/explorer-backend/testhelpers/testseed"
	"github.com/spacemeshos/explorer-backend/testhelpers/testserver"
)

const testAPIServiceDB = "explorer_test"

var (
	apiServer *testserver.TestAPIService
	generator *testseed.SeedGenerator
	dbPort    = 27017
)

func TestMain(m *testing.M) {
	mongoURL := fmt.Sprintf("mongodb://localhost:%d", dbPort)
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoURL))
	if err != nil {
		fmt.Println("failed to connect to mongo", err)
		os.Exit(1)
	}
	database := client.Database(testAPIServiceDB)
	if database != nil {
		if err = database.Drop(context.TODO()); err != nil {
			fmt.Println("failed to drop db", err)
			os.Exit(1)
		}
	}

	db, err := storage.New(context.TODO(), mongoURL, testAPIServiceDB)
	if err != nil {
		fmt.Println("failed to init storage to mongo", err)
		os.Exit(1)
	}
	seed := testseed.GetServerSeed()
	db.OnNetworkInfo(seed.NetID, 2, seed.EpochNumLayers, 4, seed.LayersDuration, 6)

	apiServer, err = testserver.StartTestAPIService(dbPort)
	if err != nil {
		fmt.Println("failed to start test api service", err)
		os.Exit(1)
	}
	apiServer.Storage = db
	generator = testseed.NewSeedGenerator(apiServer.Storage)
	if err = generator.GenerateEpoches(10); err != nil {
		fmt.Println("failed to generate epochs", err)
		os.Exit(1)
	}

	code := m.Run()
	db.Close()
	os.Exit(code)
}
