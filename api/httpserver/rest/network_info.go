package rest

import (
    "bytes"
    "net/http"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo/options"
)

func (s *Service) NetworkInfoHandler(w http.ResponseWriter, r *http.Request) {
    _ = s.process("GET", w, r, func(reqID uint64, requestBuf []byte, buf *bytes.Buffer) (Header, int, error) {

        pageNumber, pageSize, err := getPaginationInfo(r)
        if err != nil {
        }

        data, err := s.storage.GetEpochs(s.ctx, &bson.D{}, options.Find().SetSort(bson.D{{"number", -1}}).SetLimit(1).SetProjection(bson.D{{"_id", 0}}))
        if err != nil {
        }

        buf.WriteByte('{')

        setDataInfo(buf, data)
        buf.WriteByte(',')

        header := Header{}
        header["Content-Type"] = "application/json"

        err = setPaginationInfo(buf, 1, pageNumber, pageSize)
        if err != nil {
        }

        buf.WriteByte('}')

        return header, http.StatusOK, nil
    })
}
