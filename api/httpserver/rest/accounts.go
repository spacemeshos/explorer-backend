package rest

import (
    "bytes"
    "errors"
    "net/http"

    "github.com/gorilla/mux"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo/options"
)

func (s *Service) AccountsHandler(w http.ResponseWriter, r *http.Request) {
    _ = s.process("GET", w, r, func(reqID uint64, requestBuf []byte, buf *bytes.Buffer) (Header, int, error) {

        pageNumber, pageSize, err := getPaginationInfo(r)
        if err != nil {
        }

        filter := &bson.D{}

        buf.WriteByte('{')

        total := s.storage.GetAccountsCount(s.ctx, filter)
        if total > 0 {
            accounts, err := s.storage.GetAccounts(
                s.ctx,
                filter,
                options.Find().SetSort(bson.D{{"layer", -1}}).SetLimit(pageSize).SetSkip((pageNumber - 1) * pageSize).SetProjection(bson.D{
                    {"_id", 0},
                    {"layer", 0},
                }),
            )
            if err != nil {
            }
            var data []bson.D = []bson.D{}
            for _, account := range accounts {
                address := account[0].Value.(string)
                sent, received, awards, fees, timestamp := s.storage.GetAccountSummary(s.ctx, address)
                if timestamp != 0 {
                    txs := s.storage.GetTransactionsCount(s.ctx, &bson.D{
                        {"$or", bson.A{
                            bson.D{{"sender", address}},
                            bson.D{{"receiver", address}},
                        }},
                    })
                    data = append(data, bson.D{
                        account[0],
                        account[1],
                        account[2],
                        {"sent", sent},
                        {"received", received},
                        {"awards", awards},
                        {"fees", fees},
                        {"txs", txs},
                        {"timestamp", timestamp},
                    })
                } else {
                    data = append(data, bson.D{
                        account[0],
                        account[1],
                        account[2],
                        {"sent", uint64(0)},
                        {"received", uint64(0)},
                        {"awards", uint64(0)},
                        {"fees", uint64(0)},
                        {"txs", uint64(0)},
                        {"timestamp", s.storage.NetworkInfo.GenesisTime},
                    })
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

func (s *Service) AccountHandler(w http.ResponseWriter, r *http.Request) {
    _ = s.process("GET", w, r, func(reqID uint64, requestBuf []byte, buf *bytes.Buffer) (Header, int, error) {

        pageNumber, pageSize, err := getPaginationInfo(r)
        if err != nil {
        }

        vars := mux.Vars(r)
        idStr := vars["id"]
        filter := &bson.D{{"address", idStr}}

        buf.WriteByte('{')

        total := s.storage.GetAccountsCount(s.ctx, filter)
        if total > 0 {
            accounts, err := s.storage.GetAccounts(
                s.ctx,
                filter,
                options.Find().SetSort(bson.D{{"address", 1}}).SetLimit(1).SetProjection(bson.D{
                    {"_id", 0},
                    {"layer", 0},
                }),
            )
            if err != nil {
            }
            var data []bson.D = []bson.D{}
            for _, account := range accounts {
                address := account[0].Value.(string)
                sent, received, awards, fees, timestamp := s.storage.GetAccountSummary(s.ctx, address)
                if timestamp != 0 {
                    txs := s.storage.GetTransactionsCount(s.ctx, &bson.D{
                        {"$or", bson.A{
                            bson.D{{"sender", address}},
                            bson.D{{"receiver", address}},
                        }},
                    })
                    data = append(data, bson.D{
                        account[0],
                        account[1],
                        account[2],
                        {"sent", sent},
                        {"received", received},
                        {"awards", awards},
                        {"fees", fees},
                        {"txs", txs},
                        {"timestamp", timestamp},
                    })
                } else {
                    data = append(data, bson.D{
                        account[0],
                        account[1],
                        account[2],
                        {"sent", uint64(0)},
                        {"received", uint64(0)},
                        {"awards", uint64(0)},
                        {"fees", uint64(0)},
                        {"txs", uint64(0)},
                        {"timestamp", s.storage.NetworkInfo.GenesisTime},
                    })
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

func (s *Service) AccountRewardsHandler(w http.ResponseWriter, r *http.Request) {
    _ = s.process("GET", w, r, func(reqID uint64, requestBuf []byte, buf *bytes.Buffer) (Header, int, error) {

        pageNumber, pageSize, err := getPaginationInfo(r)
        if err != nil {
        }

        vars := mux.Vars(r)
        idStr := vars["id"]

        filter := &bson.D{{"coinbase", idStr}}

        buf.WriteByte('{')

        total := s.storage.GetRewardsCount(s.ctx, filter)
        if total > 0 {
            data, err := s.storage.GetRewards(s.ctx, filter, options.Find().SetSort(bson.D{{"coinbase", 1}}).SetLimit(pageSize).SetSkip((pageNumber - 1) * pageSize))
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

func (s *Service) AccountTransactionsHandler(w http.ResponseWriter, r *http.Request) {
    _ = s.process("GET", w, r, func(reqID uint64, requestBuf []byte, buf *bytes.Buffer) (Header, int, error) {

        pageNumber, pageSize, err := getPaginationInfo(r)
        if err != nil {
        }

        vars := mux.Vars(r)
        idStr := vars["id"]

        filter := &bson.D{
            {"$or", bson.A{
                bson.D{{"sender", idStr}},
                bson.D{{"receiver", idStr}},
            }},
        }

        buf.WriteByte('{')

        total := s.storage.GetTransactionsCount(s.ctx, filter)
        if total > 0 {
            data, err := s.storage.GetTransactions(s.ctx, filter, options.Find().SetSort(bson.D{{"counter", -1}}).SetLimit(pageSize).SetSkip((pageNumber - 1) * pageSize).SetProjection(bson.D{{"_id", 0}}))
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
