package blog

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func BenchmarkSigninWrapper(b *testing.B) {
	// signup a new user
	creds := Credentials{"Lucy", "123"}
	if err := creds.save(); err != nil {
		b.Fatal(err)
	}
	defer creds.remove()
	defer removeKey(creds.Username)

	handler := signinHandler
	w := httptest.NewRecorder()
	body := `{"username":"Lucy", "password":"123"}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reqbody := strings.NewReader(body)
		req := httptest.NewRequest("", "/signin", reqbody)
		handler(w, req)
	}
}
