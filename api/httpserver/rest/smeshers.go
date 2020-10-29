package rest

import (
    "bytes"
    "errors"
    "net/http"

    "github.com/gorilla/mux"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo/options"

    "github.com/spacemeshos/explorer-backend/utils"
)

func (s *Service) SmeshersHandler(w http.ResponseWriter, r *http.Request) {
    _ = s.process("GET", w, r, func(reqID uint64, requestBuf []byte, buf *bytes.Buffer) (Header, int, error) {

        pageNumber, pageSize, err := getPaginationInfo(r)
        if err != nil {
        }

        filter := &bson.D{}

        buf.WriteByte('{')

        total := s.storage.GetSmeshersCount(s.ctx, filter)
        if total > 0 {
            data, err := s.storage.GetSmeshers(s.ctx, filter, options.Find().SetSort(bson.D{{"id", 1}}).SetLimit(pageSize).SetSkip((pageNumber - 1) * pageSize).SetProjection(bson.D{{"_id", 0}}))
            if err != nil {
            }
            if data != nil && len(data) > 0 {
                for i, _ := range data {
                    smesher := utils.FindStringValue(&data[i], "id")
                    if smesher != "" {
                        rewards, _ := s.storage.GetSmesherRewards(s.ctx, smesher)
                        data[i] = append(data[i], bson.E{"rewards", rewards})
                    }
                }
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

func (s *Service) SmesherHandler(w http.ResponseWriter, r *http.Request) {
    _ = s.process("GET", w, r, func(reqID uint64, requestBuf []byte, buf *bytes.Buffer) (Header, int, error) {

        pageNumber, pageSize, err := getPaginationInfo(r)
        if err != nil {
        }

        vars := mux.Vars(r)
        idStr := vars["id"]
        filter := &bson.D{{"id", idStr}}

        buf.WriteByte('{')

        total := s.storage.GetSmeshersCount(s.ctx, filter)
        if total > 0 {
            data, err := s.storage.GetSmeshers(s.ctx, filter, options.Find().SetLimit(pageSize).SetSkip((pageNumber - 1) * pageSize).SetProjection(bson.D{{"_id", 0}}))
            if err != nil {
            }
            if data != nil && len(data) > 0 {
                for i, _ := range data {
                    smesher := utils.FindStringValue(&data[i], "id")
                    if smesher != "" {
                        rewards, _ := s.storage.GetSmesherRewards(s.ctx, smesher)
                        data[i] = append(data[i], bson.E{"rewards", rewards})
                    }
                }
            }
            setDataInfo(buf, data)
        } else {
            return nil, http.StatusNotFound, errors.New("Not found")
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

func (s *Service) SmesherRewardsHandler(w http.ResponseWriter, r *http.Request) {
    _ = s.process("GET", w, r, func(reqID uint64, requestBuf []byte, buf *bytes.Buffer) (Header, int, error) {

        pageNumber, pageSize, err := getPaginationInfo(r)
        if err != nil {
        }

        vars := mux.Vars(r)
        idStr := vars["id"]

        filter := &bson.D{{"smesher", idStr}}

        buf.WriteByte('{')

        total := s.storage.GetRewardsCount(s.ctx, filter)
        if total > 0 {
            data, err := s.storage.GetRewards(s.ctx, filter, options.Find().SetSort(bson.D{{"layer", -1}}).SetLimit(pageSize).SetSkip((pageNumber - 1) * pageSize))
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

func (s *Service) SmesherAtxsHandler(w http.ResponseWriter, r *http.Request) {
    _ = s.process("GET", w, r, func(reqID uint64, requestBuf []byte, buf *bytes.Buffer) (Header, int, error) {

        pageNumber, pageSize, err := getPaginationInfo(r)
        if err != nil {
        }

        vars := mux.Vars(r)
        idStr := vars["id"]
        filter := &bson.D{{"smesher", idStr}}

        buf.WriteByte('{')

        total := s.storage.GetActivationsCount(s.ctx, filter)
        if total > 0 {
            data, err := s.storage.GetActivations(s.ctx, filter, options.Find().SetSort(bson.D{{"id", 1}}).SetLimit(pageSize).SetSkip((pageNumber - 1) * pageSize).SetProjection(bson.D{{"_id", 0}}))
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
