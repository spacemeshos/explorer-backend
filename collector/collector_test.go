package collector_test

import (
	"context"
	"fmt"
	"github.com/spacemeshos/explorer-backend/collector"
	"github.com/spacemeshos/go-spacemesh/sql"
	"github.com/spacemeshos/go-spacemesh/sql/statesql"
	"os"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/spacemeshos/explorer-backend/storage"
	"github.com/spacemeshos/explorer-backend/test/testseed"
	"github.com/spacemeshos/explorer-backend/test/testserver"
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

	sqlDb, err := statesql.Open("file:test.db?cache=shared&mode=memory", sql.WithConnections(16),
		sql.WithMigrationsDisabled())
	seed := testseed.GetServerSeed()
	generator = testseed.NewSeedGenerator(seed)
	if err = generator.GenerateEpoches(10); err != nil {
		fmt.Println("failed to generate epochs", err)
		os.Exit(1)
	}

	dbClient := &testseed.Client{SeedGen: generator}

	node, err = testserver.CreateFakeSMNode(generator.FirstLayerTime, generator, seed)
	if err != nil {
		fmt.Println("failed to generate fake node", err)
		os.Exit(1)
	}
	defer node.Stop()
	go func() {
		if err = node.Start(); err != nil {
			fmt.Println("failed to start fake node", err)
			os.Exit(1)
		}
	}()

	privateNode, err := testserver.CreateFakeSMPrivateNode(generator.FirstLayerTime, generator, seed)
	if err != nil {
		fmt.Println("failed to generate fake private node", err)
		os.Exit(1)
	}
	defer privateNode.Stop()
	go func() {
		if err = privateNode.Start(); err != nil {
			fmt.Println("failed to start private fake node", err)
			os.Exit(1)
		}
	}()

	collectorApp = collector.NewCollector(fmt.Sprintf("localhost:%d", node.NodePort),
		fmt.Sprintf("localhost:%d", privateNode.NodePort), false,
		0, false, storageDB, sqlDb, dbClient, true)
	storageDB.AccountUpdater = collectorApp
	defer storageDB.Close()
	go collectorApp.Run()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		num := storageDB.GetRewardsCount(context.TODO(), &bson.D{})
		if int(num) == len(generator.Rewards) {
			break
		}
	}
	println("init done, start collector tests")
	code := m.Run()
	os.Exit(code)
}
