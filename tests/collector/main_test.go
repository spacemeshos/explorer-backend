package collector

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
	"testing"
	"time"

	"github.com/spacemeshos/explorer-backend/collector"
	"github.com/spacemeshos/explorer-backend/storage"
	"github.com/spacemeshos/explorer-backend/testhelpers/testseed"
	"github.com/spacemeshos/explorer-backend/testhelpers/testserver"
)

const testAPIServiceDB = "explorer_test"

var (
	dbPort       = 27017
	generator    *testseed.SeedGenerator
	node         *testserver.FakeNode
	collectorApp *collector.Collector
	storageDB    *storage.Storage
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

	storageDB, err = storage.New(context.TODO(), mongoURL, testAPIServiceDB)
	if err != nil {
		fmt.Println("failed to init storage to mongo", err)
		os.Exit(1)
	}

	seed := testseed.GetServerSeed()
	generator = testseed.NewSeedGenerator(seed)
	if err = generator.GenerateEpoches(10); err != nil {
		fmt.Println("failed to generate epochs", err)
		os.Exit(1)
	}

	node, err = testserver.CreateFakeSMNode(generator.FirstLayerTime, generator, seed)
	if err != nil {
		fmt.Println("failed to generate fake node", err)
		os.Exit(1)
	}
	go func() {
		if err = node.Start(); err != nil {
			fmt.Println("failed to start fake node", err)
			os.Exit(1)
		}
	}()

	collectorApp = collector.NewCollector(fmt.Sprintf("localhost:%d", node.NodePort), storageDB)
	storageDB.AccountUpdater = collectorApp
	go collectorApp.Run()
	time.Sleep(5 * time.Second)

	code := m.Run()
	storageDB.Close()
	if err = node.Stop(); err != nil {
		fmt.Println("failed to stop fake node", err)
		os.Exit(1)
	}
	os.Exit(code)
}
