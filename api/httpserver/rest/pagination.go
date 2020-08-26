package rest

import (
    "bytes"
    "fmt"
    "net/http"
    "strconv"
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
  pagination: {
    totalCount: 100,
    pageCount: 5,
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
    var pageNumber int64 = 1
    var pageSize int64 = 20
    var err error
    query := r.URL.Query()

    pageNumberString := query.Get("page")
    if pageNumberString != "" {
        pageNumber, err = strconv.ParseInt(pageNumberString, 10, 64)
        if err != nil {
            pageNumber = 1
        }
        if pageNumber <= 0 {
            pageNumber = 1
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

func setPaginationInfo(buf *bytes.Buffer, total int64, pageNumber int64, pageSize int64) error {
    buf.WriteString("\"pagination\":{")

    pageCount := (total + pageSize - 1) / pageSize
    buf.WriteString("\"totalCount\":")
    buf.WriteString(fmt.Sprintf("%v", total));
    buf.WriteString(",\"pageCount\":")
    buf.WriteString(fmt.Sprintf("%v", pageCount));
    buf.WriteString(",\"perPage\":")
    buf.WriteString(fmt.Sprintf("%v", pageSize));
    buf.WriteString(",\"next\":")
    if pageNumber < pageCount {
        buf.WriteString(fmt.Sprintf("%v", pageNumber + 1));
        buf.WriteString(",\"hasNext\":true")
    } else {
        buf.WriteString(fmt.Sprintf("%v", pageCount));
        buf.WriteString(",\"hasNext\":false")
    }
    buf.WriteString(",\"current\":")
    buf.WriteString(fmt.Sprintf("%v", pageNumber));
    buf.WriteString(",\"previous\":")
    if pageNumber == 1 {
        buf.WriteString("1");
        buf.WriteString(",\"hasPrevious\":false")
    } else {
        buf.WriteString(fmt.Sprintf("%v", pageNumber - 1));
        buf.WriteString(",\"hasPrevious\":true")
    }

    buf.WriteByte('}')
    return nil
}

