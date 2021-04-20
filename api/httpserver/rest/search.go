package rest

import (
    "bytes"
    "errors"
    "fmt"
    "net/http"
    "strconv"
    "strings"

    "github.com/gorilla/mux"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"

    "github.com/spacemeshos/explorer-backend/model"
)

func (s *Service) SearchHandler(w http.ResponseWriter, r *http.Request) {
    _ = s.process("GET", w, r, func(reqID uint64, requestBuf []byte, buf *bytes.Buffer) (Header, int, error) {

        buf.WriteByte('{')

        vars := mux.Vars(r)
        idStr := strings.ToLower(vars["id"])

        switch len(idStr) {
        case 42:
            address := model.ToCheckedAddress(idStr)
            if s.storage.GetAccountsCount(s.ctx, &bson.D{{"address", address}}) > 0 {
                buf.WriteString(fmt.Sprintf("\"redirect\":\"/accounts/%v\"", idStr))
                break
            }
            if s.storage.GetBlocksCount(s.ctx, &bson.D{{"id", idStr}}) > 0 {
                buf.WriteString(fmt.Sprintf("\"redirect\":\"/blocks/%v\"", idStr))
                break
            }
            return nil, http.StatusNotFound, errors.New("Not found")
        case 66:
            if s.storage.GetTransactionsCount(s.ctx, &bson.D{{"id", idStr}}) > 0 {
                buf.WriteString(fmt.Sprintf("\"redirect\":\"/txs/%v\"", idStr))
                break
            }
            if s.storage.GetActivationsCount(s.ctx, &bson.D{{"id", idStr}}) > 0 {
                buf.WriteString(fmt.Sprintf("\"redirect\":\"/atxs/%v\"", idStr))
                break
            }
            if s.storage.GetSmeshersCount(s.ctx, &bson.D{{"id", idStr}}) > 0 {
                buf.WriteString(fmt.Sprintf("\"redirect\":\"/smeshers/%v\"", idStr))
                break
            }
            if s.storage.GetLayersCount(s.ctx, &bson.D{{"hash", idStr}}) > 0 {
                layer, err := s.storage.GetLayer(s.ctx, &bson.D{{"hash", idStr}})
                if (err == nil) {
                    buf.WriteString(fmt.Sprintf("\"redirect\":\"/layers/%v\"", layer.Number))
                    break
                }
            }
            return nil, http.StatusNotFound, errors.New("Not found")
        default:
            objId, err := primitive.ObjectIDFromHex(idStr);
            if err == nil {
                if s.storage.GetRewardsCount(s.ctx, &bson.D{{"_id", objId}}) > 0 {
                    buf.WriteString(fmt.Sprintf("\"redirect\":\"/rewards/%v\"", idStr))
                    break
                }
            }
            id, err := strconv.Atoi(idStr)
            if err != nil {
                return nil, http.StatusNotFound, errors.New("Not found")
            }
            layer := s.storage.GetLastLayer(s.ctx)
            epoch := layer / s.storage.NetworkInfo.EpochNumLayers

            if uint32(id) > epoch {
                if uint32(id) <= layer {
                    buf.WriteString(fmt.Sprintf("\"redirect\":\"/layers/%v\"", id))
                } else {
                    return nil, http.StatusNotFound, errors.New("Not found")
                }
            } else {
                buf.WriteString(fmt.Sprintf("\"redirect\":\"/epochs/%v\"", id))
            }
        }

        header := Header{}
        header["Content-Type"] = "application/json"

        buf.WriteByte('}')

        return header, http.StatusOK, nil
    })
}
