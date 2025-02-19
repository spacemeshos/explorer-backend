# Explorer Stats API

Explorer Stats API is a backend service designed to provide cumulative statistics and additional data for the Spacemesh explorer frontend. It supplements API v2 of a Spacemesh node by caching data in Redis and utilizing SQLite to extract data from the node.

## Features

- Provides cumulative statistics for Spacemesh explorer frontend
- Caches frequently requested data in Redis for improved performance
- Uses SQLite to extract data from a Spacemesh node
- Complements API v2 by supplying additional metrics unavailable in the node's API

## Getting Started

### Prerequisites

- [Go](https://go.dev/) (latest version recommended)
- [Redis](https://redis.io/) for caching (optional, can use in-memory cache)
- A running Spacemesh node

### Installation

Clone the repository:

```sh
git clone https://github.com/spacemeshos/explorer-backend.git
cd explorer-backend
```

Build the project:

```sh
go build -o explorer-stats-api ./cmd/api/
```

### Configuration

Environment variables:

- `SPACEMESH_API_LISTEN`: Explorer API listen string (default: `:5000`)
- `SPACEMESH_REFRESH_API_LISTEN`: Explorer refresh API listen string (default: `:5050`)
- `SPACEMESH_TESTNET`: Enable testnet preset (`stest` instead of `sm` for wallet addresses)
- `ALLOWED_ORIGINS`: Allowed origins for CORS (default: `*`)
- `DEBUG`: Enable echo debug option along with logger middleware
- `SPACEMESH_SQLITE`: Path to node SQLite file (default: `explorer.sql`)
- `SPACEMESH_LAYERS_PER_EPOCH`: Number of layers per epoch (default: `4032`)
- `SPACEMESH_GENESIS_TIME`: Genesis time in RFC3339 format (default: `2024-06-21T13:00:00.000Z`)
- `SPACEMESH_LAYER_DURATION`: Duration of a single layer (default: `30s`)
- `SPACEMESH_LABELS_PER_UNIT`: Number of labels per unit (default: `1024`)
- `SPACEMESH_METRICS_PORT`: Metrics port (default: `:5070`)
- `SPACEMESH_CACHE_TTL`: Cache TTL for resources like overview, epochs, cumulative stats, etc. (default: `0`)
- `SPACEMESH_SHORT_CACHE_TTL`: Short Cache TTL for resources like layers, accounts, etc. (default: `5m`)
- `SPACEMESH_REDIS`: Redis address for cache; if not set, memory cache will be used

### Running the API

```sh
./explorer-stats-api
```

## API Endpoints

### General Endpoints

| Method | Endpoint               | Description                                     |
| ------ | ---------------------- | ----------------------------------------------- |
| `GET`  | `/layer/:id`           | Retrieve information about a specific layer.    |
| `GET`  | `/epoch/:id`           | Get details for a specific epoch.               |
| `GET`  | `/epoch/:id/decentral` | Retrieve decentralization metrics for an epoch. |
| `GET`  | `/account/:address`    | Fetch account details by address.               |
| `GET`  | `/smeshers/:epoch`     | List smeshers participating in a given epoch.   |
| `GET`  | `/smeshers`            | Retrieve all smeshers.                          |
| `GET`  | `/smesher/:smesherId`  | Get details of a specific smesher.              |
| `GET`  | `/overview`            | Fetch an overview of network statistics.        |
| `GET`  | `/circulation`         | Retrieve information on token circulation.      |

### Refresh Endpoints

| Method | Endpoint                       | Description                                    |
| ------ | ------------------------------ | ---------------------------------------------- |
| `GET`  | `/refresh/epoch/:id`           | Refresh cached epoch data.                     |
| `GET`  | `/refresh/epoch/:id/decentral` | Refresh decentralization metrics for an epoch. |
| `GET`  | `/refresh/overview`            | Refresh network statistics overview.           |
| `GET`  | `/refresh/smeshers/:epoch`     | Refresh smeshers list for an epoch.            |
| `GET`  | `/refresh/smeshers`            | Refresh all smeshers data.                     |
| `GET`  | `/refresh/circulation`         | Refresh token circulation data.                |

## Development

### Running in Development Mode

```sh
go run cmd/api/main.go
```
