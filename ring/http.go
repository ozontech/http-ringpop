package ring

import (
	"bytes"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"
)

// HTTPResponseWriter is a simple http.ResponseWriter implementation
type HTTPResponseWriter struct {
	headers http.Header
	body    []byte
	status  int
}

// NewResponseWriter returns new HTTP ResponseWriter
func NewResponseWriter() *HTTPResponseWriter {
	return &HTTPResponseWriter{
		headers: make(http.Header),
	}
}

func (r *HTTPResponseWriter) Header() http.Header {
	return r.headers
}

func (r *HTTPResponseWriter) Write(body []byte) (int, error) {
	r.body = append(r.body, body...)
	return len(body), nil
}

func (r *HTTPResponseWriter) WriteHeader(status int) {
	r.status = status
}

func (r *HTTPResponseWriter) Response() *http.Response {
	resp := &http.Response{
		Header:        r.headers,
		Status:        http.StatusText(r.status),
		StatusCode:    r.status,
		ProtoMajor:    1,
		ProtoMinor:    1,
		Body:          ioutil.NopCloser(bytes.NewReader(r.body)),
	}

	// Propagate Content-Length header
	if contentLengthRaw := r.Header().Get("Content-Length"); contentLengthRaw != "" {
		if contentLength, err := strconv.Atoi(contentLengthRaw); err == nil {
			resp.ContentLength = int64(contentLength)
		}
	}

	return resp
}

// RequestToKey converts request to key, that will be used for hash ring
// This func uses client IP as key
func RequestToKey(r *http.Request) string {
	for _, h := range []string{"X-Forwarded-For", "X-Real-Ip"} {
		addresses := strings.Split(r.Header.Get(h), ",")
		// march from right to left until we get a public address
		// that will be the address right before our proxy.
		for i := len(addresses) - 1; i >= 0; i-- {
			ip := strings.TrimSpace(addresses[i])
			// header can contain spaces too, strip those out.
			realIP := net.ParseIP(ip)
			if !realIP.IsGlobalUnicast() {
				// bad address, go to next
				continue
			}
			return ip
		}
	}

	return r.RemoteAddr
}
