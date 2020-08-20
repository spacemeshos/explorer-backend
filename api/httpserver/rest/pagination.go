package rest

import (
    "net/http"
    "strconv"

//    "github.com/gorilla/mux"
//    "github.com/spacemeshos/go-spacemesh/log"
//    "github.com/spacemeshos/explorer-backend/model"
//    "github.com/spacemeshos/explorer-backend/storage"
)
/*
200:
{
  data: [
    {id: 1},
    {id: 2},
    {id: 3},
    {id: 4},
    ],
  meta: {
    totalCount: 100,
    pageCount: 5
  },
  pagination: {
    perPage: 20,
    hasNext: true,
    next: 2,
    hasPrevious: false,
    current: 1,
    previous: 1
  }
} 

error:

{
  error: {
    status: 404,
    message: 'Not Found',
  }
}
*/
func getPaginationInfo(r *http.Request) (int64, int64, error) {
    var pageNumber int64
    var pageSize int64 = 20
    var err error
    query := r.URL.Query()

    pageNumberString := query.Get("page")
    if pageNumberString != "" {
        pageNumber, err = strconv.ParseInt(pageNumberString, 10, 64)
        if err != nil {
            pageNumber = 0
        }
    }
    pageSizeString := query.Get("pagesize")
    if pageSizeString != "" {
        pageSize, err = strconv.ParseInt(pageSizeString, 10, 64)
        if err != nil {
            pageSize = 20
        }
        if pageSize == 0 {
            pageSize = 20
        }
    }

    return pageNumber, pageSize, nil
}

func setPaginationInfo(buf []byte, pageNumber int64, pageSize int64) ([]byte, error) {
    return buf, nil
}
