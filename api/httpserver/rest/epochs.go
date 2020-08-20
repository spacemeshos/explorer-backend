package rest

import (
    "net/http"

//    "github.com/gorilla/mux"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo/options"

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
func (s *Service) EpochsHandler(w http.ResponseWriter, r *http.Request) {
    _ = s.process("GET", w, r, func(reqID uint64, requestBuf []byte, buf []byte) ([]byte, Header, int, error) {

        pageNumber, pageSize, err := getPaginationInfo(r)
        if err != nil {
        }

        buf = append(buf, '{')

        epochs, err := s.storage.GetEpochs(s.ctx, &bson.D{}, options.Find().SetSort(bson.D{{"number", 1}}).SetLimit(pageSize).SetSkip(pageNumber * pageSize))
        if err != nil {
        }
        if epochs == nil {
        }

//        buf = setDataInfo()

        header := Header{}

        buf, err = setPaginationInfo(buf, pageNumber, pageSize)
        if err != nil {
        }

        buf = append(buf, '}')

        return buf, header, http.StatusOK, nil
    })
}
