package rest

import (
    "bytes"
    "errors"
    "fmt"
    "net/http"
    "strconv"

    "github.com/gorilla/mux"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo/options"
)

func (s *Service) LayersHandler(w http.ResponseWriter, r *http.Request) {
    _ = s.process("GET", w, r, func(reqID uint64, requestBuf []byte, buf *bytes.Buffer) (Header, int, error) {

        pageNumber, pageSize, err := getPaginationInfo(r)
        if err != nil {
        }

        filter := &bson.D{}

        buf.WriteByte('{')

        total := s.storage.GetLayersCount(s.ctx, filter)
        if total > 0 {
            data, err := s.storage.GetLayers(s.ctx, filter, options.Find().SetSort(bson.D{{"number", -1}}).SetLimit(pageSize).SetSkip((pageNumber - 1) * pageSize).SetProjection(bson.D{{"_id", 0}}))
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

func (s *Service) LayerHandler(w http.ResponseWriter, r *http.Request) {
    _ = s.process("GET", w, r, func(reqID uint64, requestBuf []byte, buf *bytes.Buffer) (Header, int, error) {

        pageNumber, pageSize, err := getPaginationInfo(r)
        if err != nil {
        }

        vars := mux.Vars(r)
        idStr := vars["id"]
        id, err := strconv.Atoi(idStr)
        if err != nil {
            return nil, http.StatusBadRequest, fmt.Errorf("Failed to process parameter 'id' invalid number: reqID %v, id %v, error %v", reqID, idStr, err)
        }
        filter := &bson.D{{"number", id}}

        buf.WriteByte('{')

        total := s.storage.GetLayersCount(s.ctx, filter)
        if total > 0 {
            data, err := s.storage.GetLayers(s.ctx, filter, options.Find().SetLimit(1).SetProjection(bson.D{{"_id", 0}}))
            if err != nil {
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

func (s *Service) LayerTxsHandler(w http.ResponseWriter, r *http.Request) {
    _ = s.process("GET", w, r, func(reqID uint64, requestBuf []byte, buf *bytes.Buffer) (Header, int, error) {

        pageNumber, pageSize, err := getPaginationInfo(r)
        if err != nil {
        }

        vars := mux.Vars(r)
        idStr := vars["id"]
        id, err := strconv.Atoi(idStr)
        if err != nil {
            return nil, http.StatusBadRequest, fmt.Errorf("Failed to process parameter 'id' invalid number: reqID %v, id %v, error %v", reqID, idStr, err)
        }
        filter := &bson.D{{"layer", id}}

        buf.WriteByte('{')

        total := s.storage.GetTransactionsCount(s.ctx, filter)
        if total > 0 {
            data, err := s.storage.GetTransactions(s.ctx, filter, options.Find().SetSort(bson.D{{"id", 1}}).SetLimit(pageSize).SetSkip((pageNumber - 1) * pageSize).SetProjection(bson.D{{"_id", 0}}))
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

func (s *Service) LayerSmeshersHandler(w http.ResponseWriter, r *http.Request) {
    _ = s.process("GET", w, r, func(reqID uint64, requestBuf []byte, buf *bytes.Buffer) (Header, int, error) {

        pageNumber, pageSize, err := getPaginationInfo(r)
        if err != nil {
        }

        vars := mux.Vars(r)
        idStr := vars["id"]
        id, err := strconv.Atoi(idStr)
        if err != nil {
            return nil, http.StatusBadRequest, fmt.Errorf("Failed to process parameter 'id' invalid number: reqID %v, id %v, error %v", reqID, idStr, err)
        }
        filter := &bson.D{{"layer", id}}

        buf.WriteByte('{')

        total := s.storage.GetActivationsCount(s.ctx, filter)
        if total > 0 {
            data := make([]bson.D, 0)
            atxs, err := s.storage.GetActivations(s.ctx, filter, options.Find().SetSort(bson.D{{"id", 1}}).SetLimit(pageSize).SetSkip((pageNumber - 1) * pageSize).SetProjection(bson.D{
                {"_id", 0},
                {"id", 0},
                {"layer", 0},
                {"coinbase", 0},
                {"prevAtx", 0},
                {"cSize", 0},
            }))
            if err != nil {
            }
            if atxs != nil && len(atxs) > 0 {
                for _, atx := range atxs {
                    smesher, ok := atx[0].Value.(string)
                    if ok {
                        smeshers, _ := s.storage.GetSmeshers(s.ctx, &bson.D{{"id", smesher}})
                        if smeshers != nil && len(smeshers) == 1 {
                            data = append(data, smeshers[0])
                        }
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

func (s *Service) LayerBlocksHandler(w http.ResponseWriter, r *http.Request) {
    _ = s.process("GET", w, r, func(reqID uint64, requestBuf []byte, buf *bytes.Buffer) (Header, int, error) {

        pageNumber, pageSize, err := getPaginationInfo(r)
        if err != nil {
        }

        vars := mux.Vars(r)
        idStr := vars["id"]
        id, err := strconv.Atoi(idStr)
        if err != nil {
            return nil, http.StatusBadRequest, fmt.Errorf("Failed to process parameter 'id' invalid number: reqID %v, id %v, error %v", reqID, idStr, err)
        }
        filter := &bson.D{{"layer", id}}

        buf.WriteByte('{')

        total := s.storage.GetBlocksCount(s.ctx, filter)
        if total > 0 {
            data, err := s.storage.GetBlocks(s.ctx, filter, options.Find().SetSort(bson.D{{"id", 1}}).SetLimit(pageSize).SetSkip((pageNumber - 1) * pageSize).SetProjection(bson.D{{"_id", 0}}))
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

func (s *Service) LayerRewardsHandler(w http.ResponseWriter, r *http.Request) {
    _ = s.process("GET", w, r, func(reqID uint64, requestBuf []byte, buf *bytes.Buffer) (Header, int, error) {

        pageNumber, pageSize, err := getPaginationInfo(r)
        if err != nil {
        }

        vars := mux.Vars(r)
        idStr := vars["id"]
        id, err := strconv.Atoi(idStr)
        if err != nil {
            return nil, http.StatusBadRequest, fmt.Errorf("Failed to process parameter 'id' invalid number: reqID %v, id %v, error %v", reqID, idStr, err)
        }
        filter := &bson.D{{"layer", id}}

        buf.WriteByte('{')

        total := s.storage.GetRewardsCount(s.ctx, filter)
        if total > 0 {
            data, err := s.storage.GetRewards(s.ctx, filter, options.Find().SetSort(bson.D{{"smesher", 1}}).SetLimit(pageSize).SetSkip((pageNumber - 1) * pageSize))
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

func (s *Service) LayerAtxsHandler(w http.ResponseWriter, r *http.Request) {
    _ = s.process("GET", w, r, func(reqID uint64, requestBuf []byte, buf *bytes.Buffer) (Header, int, error) {

        pageNumber, pageSize, err := getPaginationInfo(r)
        if err != nil {
        }

        vars := mux.Vars(r)
        idStr := vars["id"]
        id, err := strconv.Atoi(idStr)
        if err != nil {
            return nil, http.StatusBadRequest, fmt.Errorf("Failed to process parameter 'id' invalid number: reqID %v, id %v, error %v", reqID, idStr, err)
        }
        filter := &bson.D{{"layer", id}}

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
