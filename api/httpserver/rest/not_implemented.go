package rest

import (
    "net/http"

//    "github.com/gorilla/mux"
//    "github.com/spacemeshos/go-spacemesh/log"
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
  _links: {
    self: "",
    previous: "",
    next: "",
  },
  _meta: {
    totalCount: 100,
    pageCount: 5,
    currentPage: 1,
    perPage: 20,
  },
} 

error:

{
  error: {
    status: 404,
    message: 'Not Found',
  }
}
*/


func (s *Service) NotImplemented(w http.ResponseWriter, r *http.Request) {
    _ = s.process("GET", w, r, func(reqID uint64, requestBuf []byte, buf []byte) ([]byte, Header, int, error) {

        buf = append(buf, []byte("{\"error\":{\"status\":501,\"message\":'Not Implemented'}}")...)

        return buf, nil, http.StatusNotImplemented, nil
    })
}
