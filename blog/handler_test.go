package blog

import (
	"bytes"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHttpLogger(t *testing.T) {

	// create req
	body := strings.NewReader(encodeJson(viewReq{1}))
	req := httptest.NewRequest("GET", "/viewjs", body)

	// mock writer
	w := httptest.NewRecorder()

	// logger
	b := new(bytes.Buffer)
	NewHandler().Use(HttpLogger(b)).ServeHTTP(w, req)

	got := b.String()
	want := "client request: GET /viewjs HTTP/1.1"
	if got != want {
		t.Fatalf("want: %v, but got: %v", want, got)
	}
}
