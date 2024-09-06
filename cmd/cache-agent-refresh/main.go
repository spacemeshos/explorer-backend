package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/urfave/cli/v2"
)

/* This is a simple golang app which controls the cache from the API.
 */

var (
	// Metadata.
	version string
	commit  string
	branch  string

	// All max_interval_ must be in minutes.
	max_interval_epoch_current    int = 30   // 30 minutes
	max_interval_epoch_past       int = 1440 // 1 day
	max_interval_overview         int = 60   // 1 hour
	max_interval_smeshers_current int = 30   // 30 minutes
	max_interval_smeshers_past    int = 1440 // 1 day
	max_interval_smeshers         int = 60   // 1 hour
	max_interval_circulation      int = 30   // 30 minutes

	// App settings
	// metricsPortFlag              string.
	layersPerEpoch               int
	targetNodesFlag              string // comma separated list of target nodes
	targetNodesJsonPortFlag      int
	targetNodesRefreshPortFlag   int
	targetNodesRefreshMetricFlag string
)

const (
	refresh_path_epoch              string = "/refresh/epoch/:id"
	refresh_path_epoch_decentral    string = "/refresh/epoch/:id/decentral"
	refresh_path_smeshers_per_epoch string = "/refresh/smeshers/:epoch"
	refresh_path_smeshers           string = "/refresh/smeshers"
	refresh_path_overview           string = "/refresh/overview"
	refresh_path_circulation        string = "/refresh/circulation"
)

func is_sync(latest_layer, processed_layer, tolerance int) bool {
	// Check if the node is synced
	return processed_layer >= latest_layer-tolerance
}

func get_current_epoch(leyers_per_epoch, current_layer int) int {
	// Current_Layer / Layers_Per_Epoch, if the decimal part is greater than 0.5, we are in the next epoch
	return int(math.Floor(float64(current_layer) / float64(leyers_per_epoch)))
}

func prometheus_metrics_parcer(prometheus_metric_scrape, metric_name, label_value string) float64 {
	// Get all the lines with the metric name
	// the will get the first line with contains the label value
	// will split the line by the space and get the last element

	lines := make([]string, 10)
	for _, line := range strings.Split(prometheus_metric_scrape, "\n") {
		if strings.Contains(line, metric_name) && strings.Contains(line, label_value) {
			lines = append(lines, line)
		}
	}

	if len(lines) == 0 {
		return 0
	}

	// Get the last element of the line
	last_line := lines[len(lines)-1]
	splited_line := strings.Split(last_line, " ")
	value := splited_line[len(splited_line)-1]

	result, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0
	}
	return result
}

type NodeStatus struct {
	ProcessedLayer int `json:"processed_layer"`
	LatestLayer    int `json:"latest_layer"`
}

func get_status(node string, port int) (status NodeStatus, err error) {
	url := fmt.Sprintf("%s:%d/spacemesh.v2alpha1.NodeService/Status", node, port)

	// HTTP POST request to check status
	resp, err := http.Post(url, "application/json", nil)
	if err != nil || resp.StatusCode != 200 {
		fmt.Printf("Error checking node %s:%d, status code: %d\n", node, port, resp.StatusCode)
		return
	}
	defer resp.Body.Close()

	// Read and parse response JSON
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response from node %s:%d\n", node, port)
		return
	}

	// Parse the response into NodeStatus struct

	err = json.Unmarshal(body, &status)
	if err != nil {
		fmt.Printf("Error parsing JSON response from node %s:%d\n", node, port)
		return
	}

	return
}

// checkNodeStatus checks the status of a node and whether it's synced.
func checkNodeStatus(node string, port int) (bool, string) {
	// Parse the response into NodeStatus struct
	status, err := get_status(node, port)
	if err != nil {
		fmt.Printf("Error getting status from node %s:%d\n", node, port)
		return false, ""
	}

	// Check if node is synced using the is_sync function with a tolerance of 2 layers
	if is_sync(status.LatestLayer, status.ProcessedLayer, 2) {
		fmt.Printf("Node %s:%d is online and synced\n", node, port)
		return true, node
	}
	return false, ""
}

func refresh_cache(node string, port int, path string, interval int, prometheus string) error {
	log.Printf("Refreshing cache for %s:%d%s\n", node, port, path)

	last_refresh := prometheus_metrics_parcer(prometheus, "cache_agent_last_refresh", path)
	if last_refresh < float64(interval*60) {
		return nil
	}

	url := fmt.Sprintf("%s:%d%s", node, port, path)

	// HTTP GET request to refresh cache
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != 200 {
		fmt.Printf("Error refreshing cache for %s:%d%s, status code: %d\n", node, port, path, resp.StatusCode)
		return err
	}
	defer resp.Body.Close()

	log.Printf("Cache refreshed for %s:%d%s\n", node, port, path)
	return nil
}

func epoch_replace(path string, epoch int) string {
	// Replace the :id and :epoch with the current epoch
	epoch_path := strings.ReplaceAll(path, ":id", strconv.Itoa(epoch))
	epoch_path = strings.ReplaceAll(epoch_path, ":epoch", strconv.Itoa(epoch))
	return epoch_path
}

var flags = []cli.Flag{
	&cli.IntFlag{
		Name:        "layers-per-epoch",
		Usage:       "Number of layers per epoch",
		Required:    false,
		Destination: &layersPerEpoch,
		Value:       4032,
		EnvVars:     []string{"SPACEMESH_LAYERS_PER_EPOCH"},
	},
	&cli.StringFlag{
		Name:        "target-nodes",
		Usage:       "Comma separated list of target nodes",
		Required:    true,
		Destination: &targetNodesFlag,
		EnvVars:     []string{"TARGET_NODES"},
	},
	&cli.IntFlag{
		Name:        "target-nodes-json-port",
		Usage:       "Port for the JSON API of the target nodes",
		Required:    true,
		Destination: &targetNodesJsonPortFlag,
		EnvVars:     []string{"TARGET_NODES_JSON_PORT"},
	},
	&cli.IntFlag{
		Name:        "target-nodes-refresh-port",
		Usage:       "Port for the refresh API of the target nodes",
		Required:    true,
		Destination: &targetNodesRefreshPortFlag,
		EnvVars:     []string{"TARGET_NODES_REFRESH_PORT"},
	},
	&cli.StringFlag{
		Name:        "target-nodes-refresh-metric",
		Usage:       "Port for the refresh API of the target nodes",
		Required:    true,
		Destination: &targetNodesRefreshMetricFlag,
		EnvVars:     []string{"TARGET_NODES_REFRESH_METRIC"},
	},
}

func main() {
	app := cli.NewApp()
	app.Name = "cache-agent-refresh"
	app.Version = fmt.Sprintf("%s, commit '%s', branch '%s'", version, commit, branch)
	app.Flags = flags
	app.Writer = os.Stderr

	app.Action = func(ctx *cli.Context) error {
		nodes := strings.Split(targetNodesFlag, ",")
		var targetNode string
		// Check if the nodes are synced
		for _, node := range nodes {
			fmt.Printf("Checking node %s\n", node)
			synced, _ := checkNodeStatus(node, targetNodesJsonPortFlag)
			if synced {
				targetNode = node
				break
			}
		}

		if targetNode == "" {
			fmt.Println("No synced nodes found")
			return errors.New("no synced nodes found")
		}

		// Get the prometheus metrics
		prometheus_url := fmt.Sprintf("http://%s:%s/metrics", targetNode, targetNodesRefreshMetricFlag)
		resp, err := http.Get(prometheus_url)
		if err != nil {
			log.Fatalf("Error getting prometheus metrics: %v", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("Error reading prometheus metrics: %v", err)
		}

		// Get the current epoch
		nodeStatus, err := get_status(targetNode, targetNodesJsonPortFlag)
		if err != nil {
			log.Fatalf("Error getting node status: %v", err)
		}
		current_epoch := get_current_epoch(layersPerEpoch, nodeStatus.LatestLayer)

		// refresh the cache of  smeshers, overview and circulation
		wg := sync.WaitGroup{}
		wg.Add(3)
		go func() {
			defer wg.Done()
			refresh_cache(targetNode, targetNodesRefreshPortFlag, refresh_path_overview, max_interval_overview,
				string(body))
		}()

		go func() {
			defer wg.Done()
			refresh_cache(targetNode, targetNodesRefreshPortFlag, refresh_path_circulation, max_interval_circulation,
				string(body))
		}()

		go func() {
			defer wg.Done()
			refresh_cache(targetNode, targetNodesRefreshPortFlag, refresh_path_smeshers, max_interval_smeshers,
				string(body))
		}()

		wg.Wait()

		wg.Add(3)

		go func() {
			defer wg.Done()
			path := epoch_replace(refresh_path_epoch, current_epoch)
			refresh_cache(targetNode, targetNodesRefreshPortFlag, path, max_interval_epoch_current, string(body))
		}()

		go func() {
			defer wg.Done()
			path := epoch_replace(refresh_path_epoch_decentral, current_epoch)
			refresh_cache(targetNode, targetNodesRefreshPortFlag, path, max_interval_epoch_current, string(body))
		}()

		go func() {
			defer wg.Done()
			path := epoch_replace(refresh_path_smeshers_per_epoch, current_epoch)
			refresh_cache(targetNode, targetNodesRefreshPortFlag, path, max_interval_smeshers_current, string(body))
		}()

		wg.Wait()

		// refresh the cache of the past epoch
		for epoch := current_epoch - 1; epoch >= 0; epoch-- {
			wg.Add(3)
			go func(epoch int) {
				defer wg.Done()
				path := epoch_replace(refresh_path_epoch, epoch)
				refresh_cache(targetNode, targetNodesRefreshPortFlag, path, max_interval_epoch_past, string(body))
			}(epoch)

			go func(epoch int) {
				defer wg.Done()
				path := epoch_replace(refresh_path_epoch_decentral, epoch)
				refresh_cache(targetNode, targetNodesRefreshPortFlag, path, max_interval_epoch_past, string(body))
			}(epoch)

			go func(epoch int) {
				defer wg.Done()
				path := epoch_replace(refresh_path_smeshers_per_epoch, epoch)
				refresh_cache(targetNode, targetNodesRefreshPortFlag, path, max_interval_smeshers_past, string(body))
			}(epoch)

			wg.Wait()
		}
		return nil
	}

	log.Println("Cache refresh completed")
	os.Exit(0)
}
