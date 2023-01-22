package blog

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func BenchmarkAnalyzeHandler(b *testing.B) {

	reqbody := `{"how": 2, "id": 1}`
	w := httptest.NewRecorder()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {

		req := httptest.NewRequest("", "/analyze", strings.NewReader(reqbody))
		req.Header.Set("Cookie", cookie)
		analyzeHandler(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			b.Fatalf("code %v, body %v", resp.StatusCode, string(body))
		}
	}
}
