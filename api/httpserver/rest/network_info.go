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

		networkInfo, err := s.storage.GetNetworkInfo(s.ctx)
		if err == nil {
			s.storage.NetworkInfo.LastLayer = networkInfo.LastLayer
			s.storage.NetworkInfo.LastLayerTimestamp = networkInfo.LastLayerTimestamp
			s.storage.NetworkInfo.LastApprovedLayer = networkInfo.LastApprovedLayer
			s.storage.NetworkInfo.LastConfirmedLayer = networkInfo.LastConfirmedLayer
			s.storage.NetworkInfo.ConnectedPeers = networkInfo.ConnectedPeers
			s.storage.NetworkInfo.IsSynced = networkInfo.IsSynced
			s.storage.NetworkInfo.SyncedLayer = networkInfo.SyncedLayer
			s.storage.NetworkInfo.TopLayer = networkInfo.TopLayer
			s.storage.NetworkInfo.VerifiedLayer = networkInfo.VerifiedLayer
		}

		buf.WriteString("\"network\":")
		writeD(buf, &bson.D{
			{"netid", s.storage.NetworkInfo.GenesisId},
			{"genesis", s.storage.NetworkInfo.GenesisTime},
			{"layers", s.storage.NetworkInfo.EpochNumLayers},
			{"maxtx", s.storage.NetworkInfo.MaxTransactionsPerSecond},
			{"duration", s.storage.NetworkInfo.LayerDuration},
			{"lastlayer", s.storage.NetworkInfo.LastLayer},
			{"lastlayerts", s.storage.NetworkInfo.LastLayerTimestamp},
			{"lastapprovedlayer", s.storage.NetworkInfo.LastApprovedLayer},
			{"lastconfirmedlayer", s.storage.NetworkInfo.LastConfirmedLayer},
			{"connectedpeers", s.storage.NetworkInfo.ConnectedPeers},
			{"issynced", s.storage.NetworkInfo.IsSynced},
			{"syncedlayer", s.storage.NetworkInfo.SyncedLayer},
			{"toplayer", s.storage.NetworkInfo.TopLayer},
			{"verifiedlayer", s.storage.NetworkInfo.VerifiedLayer},
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

func (s *Service) HealthzHandler(w http.ResponseWriter, r *http.Request) {
	_ = s.process("GET", w, r, func(reqID uint64, requestBuf []byte, buf *bytes.Buffer) (Header, int, error) {

		header := Header{}
		header["Content-Type"] = "plain/text"

		err := s.Ping()
		if err == nil {
			buf.WriteString("OK")
			return header, http.StatusOK, nil
		}

		return header, http.StatusServiceUnavailable, nil
	})
}

func (s *Service) SyncedHandler(w http.ResponseWriter, r *http.Request) {
	_ = s.process("GET", w, r, func(reqID uint64, requestBuf []byte, buf *bytes.Buffer) (Header, int, error) {

		status := http.StatusTooEarly

		networkInfo, err := s.storage.GetNetworkInfo(s.ctx)
		if err == nil && networkInfo.IsSynced {
			status = http.StatusOK
			buf.WriteString("SYNCED")
		} else {
			buf.WriteString("SYNCING")
		}

		header := Header{}
		header["Content-Type"] = "text/plain"

		return header, status, nil
	})
}
