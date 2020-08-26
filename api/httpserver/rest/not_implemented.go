package rest

import (
    "bytes"
    "net/http"
)

func (s *Service) NotImplemented(w http.ResponseWriter, r *http.Request) {
    _ = s.process("GET", w, r, func(reqID uint64, requestBuf []byte, buf *bytes.Buffer) (Header, int, error) {

        buf.WriteString("{\"error\":{\"status\":501,\"message\":'Not Implemented'}}")

        return nil, http.StatusNotImplemented, nil
    })
}
