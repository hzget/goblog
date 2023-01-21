package blog

import (
	"testing"
	"net/http/httptest"
	"strings"
)

func BenchmarkViewHandler(b *testing.B) {
	handler := makeHandler(viewHandler)
	w := httptest.NewRecorder()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/view/1", nil)
		req.Header.Set("Cookie", cookie)
		handler(w, req)
	}
}

func BenchmarkViewjs(b *testing.B) {
	// create req
	handler := makePageHandler(viewjsHandler)

	w := httptest.NewRecorder()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		body := strings.NewReader(encodeJson(viewReq{1}))
		req := httptest.NewRequest("POST", "/viewjs, body)
		req.Header.Set("Cookie", cookie)
		handler(w, req)
	}

}

