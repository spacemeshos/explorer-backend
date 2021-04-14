# Explorer-backend
Spacemesh explorer backend designed to provide data for explorer-frontends

## Explorer Software Architecture
![](https://raw.githubusercontent.com/spacemeshos/product/master/resources/explorer_arch_chart.png)

## Using the Explorer Backend API
The explorer backend provides a public REST API that can be used to get data about a Spacemesh network.
Follow these steps to use the API for a public Spacemesh network:

1. Obtain a currently available explorer API endpoint from the [Spacemesh public web services](https://discover.spacemesh.io/networks.json) endpoint. This endpoint lists all available Spacemesh networks such as testnets.
1. Build a REST request using the endpoint. For example, if the explorer api url for net-id 28 is `https://explorer-api-28.spacemesh.io/` then the network-info data is available at `https://explorer-api-28.spacemesh.io/network-info`.
1. Issue an http 'GET' request to get the data. e.g. `curl https://explorer-api-28.spacemesh.io/network-info`. 
1. Live long and prosper.

### Paging and pagination
- Use the `pagesize` and `page` params to get paginated results. The first page number is 1, so for example, to get the first 20 accounts on TN 128 call: `https://explorer-api-28.spacemesh.io/accounts?pagesize=20&page=1` and to get the next 20 accounts use: `https://explorer-api-28.spacemesh.io/accounts?pagesize=20&page=2`
- API results which support pagination include pagination data in the response json. e.g.:

```
{"totalCount":1020,"pageCount":51,"perPage":20,"next":2,"hasNext":true,"current":1,"previous":1,"hasPrevious":false}}
```

Use this pagination data to figure out how many calls you need to make and which what params in order to get all the data.


### API Capabilities
The API is not properly documented yet. The best way to identity the supported API methods is via the api server [source code](https://github.com/spacemeshos/explorer-backend/blob/master/api/httpserver/httpserver.go).

### API Usage Examples

- Get an account state: https://explorer-api-28.spacemesh.io/accounts/0xaFEd9A1c17Ca7eaA7A6795dBc7BEe1B1d992c7ba



