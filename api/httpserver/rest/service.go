package rest

import (
    "bytes"
    "context"
    "fmt"
    "errors"
    "io/ioutil"
    "net/http"
    "sync"
    "sync/atomic"

    "github.com/spacemeshos/go-spacemesh/log"
    "github.com/spacemeshos/explorer-backend/storage"
)

type Header map[string]string

type Service struct {
    ctx      context.Context
    cancel   context.CancelFunc
    storage  *storage.Storage

    pool sync.Pool
}

var requestID uint64

func GetNextRequestID() uint64 {
    return atomic.AddUint64(&requestID, 1)
}

func New(ctx context.Context, storage *storage.Storage) (*Service, error) {

    log.Info("Creating new REST service")

    service := &Service{
        storage: storage,
        pool: sync.Pool{
            New: func() interface{} {
                return make([]byte, 4096)
            },
        },
    }

    if ctx == nil {
        service.ctx, service.cancel = context.WithCancel(context.Background())
    } else {
        service.ctx, service.cancel = context.WithCancel(ctx)
    }

    log.Info("REST service is created")
    return service, nil
}

func (s *Service) Shutdown() error {
    defer s.cancel()

    return nil
}

func (s *Service) process(method string, w http.ResponseWriter, r *http.Request, fn func(reqID uint64, requestBuf []byte, buf *bytes.Buffer) (Header, int, error)) error {

    reqId := GetNextRequestID()

    log.Info("Check allowed HTTP method: %v, %v", r.Method, method)
    if r.Method != method {
        return fmt.Errorf("HTTP method is not allowed: %v, request.Method %v, method %v", reqId, r.Method, method)
    }

    log.Info("Reading request body: %v", reqId)
    requestBuf, err := ioutil.ReadAll(r.Body)
    if err != nil {
        return fmt.Errorf("Failed to read HTTP body: reqID %v, err %v", reqId, err)
    }
    log.Info("Read request body: reqID %v, len(body) %v", reqId, len(requestBuf))

    var responseBuf bytes.Buffer

    header, status, err := fn(reqId, requestBuf, &responseBuf)
    if err != nil {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.WriteHeader(status)
        log.Info("Process errro %v", err)
        return err
    }

    if header != nil {
        for key, h := range header {
            w.Header().Set(key, h)
        }
    }

    w.Header().Set("Access-Control-Allow-Origin", "*")

    log.Info("Set HTTP response status: reqID %v, Status %v", reqId, http.StatusText(status))
    w.WriteHeader(status)

    if responseBuf.Len() > 0 {
        log.Info("Writing HTTP response body: reqID %v, len(body) %v", reqId, responseBuf.Len())
        respWrittenLen, err := responseBuf.WriteTo(w)
        if err != nil {
            log.Info("Failed to write HTTP repsonse: reqID %v error %v", reqId, err)
            return err
        }
        log.Info("Written HTTP response: reqID %v, len(body) %v", reqId, respWrittenLen)
    } else {
        log.Info("HTTP response body is empty")
    }

    return nil
}

func (s *Service) Ping() error {
    if s.storage == nil {
        return errors.New("Serice not initialized")
    }
    return s.storage.Ping()
}
