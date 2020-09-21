package rest

import (
    "bytes"
    "errors"
    "fmt"
    "net/http"
    "strconv"

    "github.com/spacemeshos/go-spacemesh/log"

    "github.com/gorilla/mux"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo/options"
)

func (s *Service) EpochsHandler(w http.ResponseWriter, r *http.Request) {
    _ = s.process("GET", w, r, func(reqID uint64, requestBuf []byte, buf *bytes.Buffer) (Header, int, error) {

        pageNumber, pageSize, err := getPaginationInfo(r)
        if err != nil {
        }

        filter := &bson.D{}

        buf.WriteByte('{')

        total := s.storage.GetEpochsCount(s.ctx, filter)
        if total > 0 {
            epochs, err := s.storage.GetEpochs(s.ctx, &bson.D{}, options.Find().SetSort(bson.D{{"number", -1}}).SetLimit(pageSize).SetSkip((pageNumber - 1) * pageSize).SetProjection(bson.D{{"_id", 0}}))
            if err != nil {
            }
            setDataInfo(buf, epochs)
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

func (s *Service) EpochHandler(w http.ResponseWriter, r *http.Request) {
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

        total := s.storage.GetEpochsCount(s.ctx, filter)
        if total > 0 {
            data, err := s.storage.GetEpochs(s.ctx, filter, options.Find().SetLimit(1).SetProjection(bson.D{{"_id", 0}}))
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

func (s *Service) EpochTxsHandler(w http.ResponseWriter, r *http.Request) {
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
        layerStart, layerEnd := s.storage.GetEpochLayers(int32(id))
        filter := &bson.D{{"layer", bson.D{{"$gte", layerStart}, {"$lte", layerEnd}}}}

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

func (s *Service) EpochSmeshersHandler(w http.ResponseWriter, r *http.Request) {
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
        layerStart, layerEnd := s.storage.GetEpochLayers(int32(id))
        filter := &bson.D{{"layer", bson.D{{"$gte", layerStart}, {"$lte", layerEnd}}}}

        buf.WriteByte('{')

        atxs, err := s.storage.GetActivations(s.ctx, filter, options.Find().SetSort(bson.D{{"id", 1}}).SetLimit(pageSize).SetSkip((pageNumber - 1) * pageSize).SetProjection(bson.D{
            {"_id", 0},
            {"id", 0},
            {"layer", 0},
            {"coinbase", 0},
            {"prevAtx", 0},
            {"cSize", 0},
        }))
        smeshers := make([]string, 0, len(atxs))
        var lastId string
        for _, atx := range atxs {
            smesherId := atx[0].Value.(string)
            if lastId != smesherId {
                smeshers = append(smeshers, smesherId)
                lastId = smesherId
            }
        }

        total := int64(len(smeshers))
        var dataSet bool
        if total > 0 {
            from := (pageNumber - 1) * pageSize
            if from < total {
                to := from + pageSize
                if to > total {
                    to = total
                }
                var data []bson.D = []bson.D{}
                for _, smesherId := range smeshers {
                    smesher, err := s.storage.GetSmesher(s.ctx, &bson.D{{"id", smesherId}})
                    if err != nil {
                    }
                    data = append(data, bson.D{
                        {"id", smesher.Id},
                        {"name", smesher.Geo.Name},
                        {"lon", smesher.Geo.Coordinates[0]},
                        {"lat", smesher.Geo.Coordinates[1]},
                        {"cSize", smesher.CommitmentSize},
                    })
                }
                setDataInfo(buf, data)
                dataSet = true
            }
        }
        if !dataSet {
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

func (s *Service) EpochLayersHandler(w http.ResponseWriter, r *http.Request) {
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
        layerStart, layerEnd := s.storage.GetEpochLayers(int32(id))
        filter := &bson.D{{"number", bson.D{{"$gte", layerStart}, {"$lte", layerEnd}}}}

        buf.WriteByte('{')

        total := s.storage.GetLayersCount(s.ctx, filter)
        log.Info("EpochLayersHandler: %v-%v, filter: %+v, total: %v", layerStart, layerEnd, filter, total)
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

func (s *Service) EpochRewardsHandler(w http.ResponseWriter, r *http.Request) {
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
        layerStart, layerEnd := s.storage.GetEpochLayers(int32(id))
        filter := &bson.D{{"layer", bson.D{{"$gte", layerStart}, {"$lte", layerEnd}}}}

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

func (s *Service) EpochAtxsHandler(w http.ResponseWriter, r *http.Request) {
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
        layerStart, layerEnd := s.storage.GetEpochLayers(int32(id))
        filter := &bson.D{{"layer", bson.D{{"$gte", layerStart}, {"$lte", layerEnd}}}}

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
