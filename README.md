# Explorer-backend
Spacemesh explorer backend designed to provide data for explorer-frontends

## Explorer Software Architecture
![](https://raw.githubusercontent.com/spacemeshos/product/master/resources/explorer_arch_chart.png)

## Using the Explorer Backend API
The explorer backend provides a public REST API that can be used to get data about a Spacemesh network.
Follow these steps to use the API for a public Spacemesh network:

1. Obtain a currently available explorer API endpoint from the [Spacemesh public web services](https://configs.spacemesh.network/networks.json) endpoint. This endpoint lists all available Spacemesh networks.
1. Build a REST request using the endpoint. For example, explorer api url for mainnet is `https://mainnet-explorer-api.spacemesh.network/` then the network-info data is available at `https://mainnet-explorer-api.spacemesh.network/network-info`.
1. Issue an http 'GET' request to get the data. e.g. `curl https://mainnet-explorer-api.spacemesh.network/network-info`. 
1. Live long and prosper.

### Paging and pagination
- Use the `pagesize` and `page` params to get paginated results. The first page number is 1, so for example, to get the first 20 layers call: `https://mainnet-explorer-api.spacemesh.network/layers?pagesize=20&page=1` and to get the next 20 layers use: `https://mainnet-explorer-api.spacemesh.network/layers?pagesize=20&page=2`
- API results which support pagination include pagination data in the response json. e.g.:

```
{"totalCount":1020,"pageCount":51,"perPage":20,"next":2,"hasNext":true,"current":1,"previous":1,"hasPrevious":false}}
```

Use this pagination data to figure out how many calls you need to make and which what params in order to get all the data.


### API Capabilities
The API is not properly documented yet. The best way to identity the supported API methods is via the api server [source code](https://github.com/spacemeshos/explorer-backend/blob/master/internal/api/router/router.go).

### API Usage Examples

- Get layer details: https://mainnet-explorer-api.spacemesh.network/layers/52410
- Get mainnet current network info: https://mainnet-explorer-api.spacemesh.network/


### Explorer Stats API

```shell
GLOBAL OPTIONS:
   --listen value                                       Explorer API listen string in format <host>:<port> (default: ":5000") [$SPACEMESH_API_LISTEN]
   --listen-refresh value                               Explorer refresh API listen string in format <host>:<port> (default: ":5050") [$SPACEMESH_REFRESH_API_LISTEN]
   --testnet                                            Use this flag to enable testnet preset ("stest" instead of "sm" for wallet addresses) (default: false) [$SPACEMESH_TESTNET]
   --allowed-origins value [ --allowed-origins value ]  Use this flag to set allowed origins for CORS (default: "*") [$ALLOWED_ORIGINS]
   --debug                                              Use this flag to enable echo debug option along with logger middleware (default: false) [$DEBUG]
   --sqlite value                                       Path to node sqlite file (default: "explorer.sql") [$SPACEMESH_SQLITE]
   --layers-per-epoch value                             Number of layers per epoch (default: 4032) [$SPACEMESH_LAYERS_PER_EPOCH]
   --genesis-time value                                 Genesis time in RFC3339 format (default: "2024-06-21T13:00:00.000Z") [$SPACEMESH_GENESIS_TIME]
   --layer-duration value                               Duration of a single layer (default: 30s) [$SPACEMESH_LAYER_DURATION]
   --labels-per-unit value                              Number of labels per unit (default: 1024) [$SPACEMESH_LABELS_PER_UNIT]
   --bits-per-label value                               Number of bits per label (default: 128) [$SPACEMESH_BITS_PER_LABEL]
   --metricsPort value                                  (default: ":5070") [$SPACEMESH_METRICS_PORT]
   --help, -h                                           show help
   --version, -v                                        print the version

```