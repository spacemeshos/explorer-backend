package rest

import (
    "bytes"
//    "errors"
    "net/http"

    "github.com/gorilla/mux"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo/options"
)

func (s *Service) AppsHandler(w http.ResponseWriter, r *http.Request) {
    _ = s.process("GET", w, r, func(reqID uint64, requestBuf []byte, buf *bytes.Buffer) (Header, int, error) {

        pageNumber, pageSize, err := getPaginationInfo(r)
        if err != nil {
        }

        filter := &bson.D{}

        buf.WriteByte('{')

        total := s.storage.GetAppsCount(s.ctx, filter)
        if total > 0 {
            data, err := s.storage.GetApps(s.ctx, filter, options.Find().SetSort(bson.D{{"address", 1}}).SetLimit(pageSize).SetSkip((pageNumber - 1) * pageSize).SetProjection(bson.D{{"_id", 0}}))
            if err != nil {
            }
            setDataInfo(buf, data)
        } else {
            setDataInfo(buf, nil)
        }

        buf.WriteByte(',')

        header := Header{}
        header["Content-Type"] = "application/json"

        err = setPaginationInfo(buf, total, pageNumber, pageSize)
        if err != nil {
        }

        buf.WriteByte('}')

        return header, http.StatusOK, nil
    })
}

func (s *Service) AppHandler(w http.ResponseWriter, r *http.Request) {
    _ = s.process("GET", w, r, func(reqID uint64, requestBuf []byte, buf *bytes.Buffer) (Header, int, error) {

        pageNumber, pageSize, err := getPaginationInfo(r)
        if err != nil {
        }

        vars := mux.Vars(r)
        idStr := vars["id"]
        filter := &bson.D{{"address", idStr}}

        buf.WriteByte('{')

        total := s.storage.GetAppsCount(s.ctx, filter)
        if total > 0 {
            data, err := s.storage.GetApps(s.ctx, filter, options.Find().SetLimit(1).SetProjection(bson.D{{"_id", 0}}))
            if err != nil {
            }
            setDataInfo(buf, data)
        } else {
            setDataInfo(buf, nil)
//            return nil, http.StatusNotFound, errors.New("Not found")
        }

        buf.WriteByte(',')

        header := Header{}
        header["Content-Type"] = "application/json"

        err = setPaginationInfo(buf, total, pageNumber, pageSize)
        if err != nil {
        }

        buf.WriteByte('}')

        return header, http.StatusOK, nil
    })
}
