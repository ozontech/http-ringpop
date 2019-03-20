package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCopyHTTPResponseFromRaw(t *testing.T) {
	var rawBackendResponse = []byte(`HTTP/1.1 201 OK
Content-Length: 50
Content-Type: text/plain; charset=utf-8
Date: Tue, 06 Nov 2018 20:59:14 GMT
X-Proxy: ringpop

Hello from backend :4001 to client 127.0.0.1:50245`)

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "http://localhost/", nil)

	if err := copyHTTPResponseFromRaw(w, r, rawBackendResponse); err != nil {
		t.Fatalf("Error on coping HTTP response: %v", err)
	}

	if w.Code != 201 {
		t.Fatalf("Unexpecetd response code: %d, expected: 201", w.Code)
	}

	expectedBody := `Hello from backend :4001 to client 127.0.0.1:50245`
	if w.Body.String() != expectedBody {
		t.Fatalf("Unexpected response body: \n%s\nexpected:\n%s", w.Body.String(), expectedBody)
	}
}

func TestHTTPRequestToBytes(t *testing.T) {
	r, _ := http.NewRequest("GET", "http://localhost/", nil)
	rBytes, err := httpRequestToBytes(r)

	if err != nil {
		t.Fatalf("Error on converting HTTP request to bytes: %v", err)
	}

	expectedRequest := []byte{71, 69, 84, 32, 47, 32, 72, 84, 84, 80, 47, 49, 46, 49, 13, 10, 72, 111, 115, 116, 58, 32, 108, 111, 99, 97, 108, 104, 111, 115, 116, 13, 10, 85, 115, 101, 114, 45, 65, 103, 101, 110, 116, 58, 32, 71, 111, 45, 104, 116, 116, 112, 45, 99, 108, 105, 101, 110, 116, 47, 49, 46, 49, 13, 10, 13, 10}

	if string(rBytes) != string(expectedRequest) {
		t.Fatalf("Requests are not equal, got: \n%s\n, expected: \n%s", string(rBytes), string(expectedRequest))
	}
}

