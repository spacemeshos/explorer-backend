package rest

import (
    "bytes"
    "net/http"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo/options"
)

func (s *Service) NetworkInfoHandler(w http.ResponseWriter, r *http.Request) {
    _ = s.process("GET", w, r, func(reqID uint64, requestBuf []byte, buf *bytes.Buffer) (Header, int, error) {

        buf.WriteByte('{')

        buf.WriteString("\"network\":")
        writeD(buf, &bson.D{
            {"netid", s.storage.NetworkInfo.NetId},
            {"genesis", s.storage.NetworkInfo.GenesisTime},
            {"layers", s.storage.NetworkInfo.EpochNumLayers},
            {"maxtx", s.storage.NetworkInfo.MaxTransactionsPerSecond},
            {"duration", s.storage.NetworkInfo.LayerDuration},
            {"lastlayer", s.storage.NetworkInfo.LastLayer},
            {"lastlayerts", s.storage.NetworkInfo.LastLayerTimestamp},
            {"lastapprovedlayer", s.storage.NetworkInfo.LastApprovedLayer},
            {"lastconfirmedlayer", s.storage.NetworkInfo.LastConfirmedLayer},
        })

        epoch, err := s.storage.GetEpochs(s.ctx, &bson.D{}, options.Find().SetSort(bson.D{{"number", -1}}).SetLimit(1).SetProjection(bson.D{{"_id", 0}}))
        if err != nil {
        }
        if epoch != nil && len(epoch) == 1 {
            buf.WriteString(",\"epoch\":")
            writeD(buf, &epoch[0])
        }

        layer, err := s.storage.GetLayers(s.ctx, &bson.D{}, options.Find().SetSort(bson.D{{"number", -1}}).SetLimit(1).SetProjection(bson.D{{"_id", 0}}))
        if err != nil {
        }
        if layer != nil && len(layer) == 1 {
            buf.WriteString(",\"layer\":")
            writeD(buf, &layer[0])
        }

        header := Header{}
        header["Content-Type"] = "application/json"

        buf.WriteByte('}')

        return header, http.StatusOK, nil
    })
}
