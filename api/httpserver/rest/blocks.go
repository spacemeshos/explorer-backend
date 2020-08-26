package rest

import (
    "bytes"
    "net/http"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo/options"
)

func (s *Service) BlocksHandler(w http.ResponseWriter, r *http.Request) {
    _ = s.process("GET", w, r, func(reqID uint64, requestBuf []byte, buf *bytes.Buffer) (Header, int, error) {

        pageNumber, pageSize, err := getPaginationInfo(r)
        if err != nil {
        }

        filter := &bson.D{}

        total, err := s.storage.GetBlocksCount(s.ctx, filter)
        if err != nil {
        }

        data, err := s.storage.GetBlocks(s.ctx, filter, options.Find().SetSort(bson.D{{"id", 1}}).SetLimit(pageSize).SetSkip((pageNumber - 1) * pageSize).SetProjection(bson.D{{"_id", 0}}))
        if err != nil {
        }

        buf.WriteByte('{')

        setDataInfo(buf, data)
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
