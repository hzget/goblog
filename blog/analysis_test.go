package blog

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAnalyzeHandler(t *testing.T) {
	cases := []struct {
		name string
		req  string
		code int
		want jsonResp
	}{
		{"Normal", `{"how":2, "id":1}`, http.StatusOK, jsonResp{true, "Sports"}},
		{"Notlogin", `{"how":2, "id":1}`, http.StatusUnauthorized, jsonResp{false, "please log in first"}},
		{"MalformedJson", `{"how:2, "id":1}`, http.StatusBadRequest, jsonResp{false, ""}},
		{"Unkownfield", `{"howe":2, "id":1}`, http.StatusBadRequest, jsonResp{false, ""}},
		{"InvalidId(-1)", `{"how":2, "id":-1}`, http.StatusBadRequest, jsonResp{false, ""}},
		{"InvalidId(0)", `{"how":2, "id":0}`, http.StatusBadRequest, jsonResp{false, ""}},
		{"BadRequestMethod", `{"how":0, "id":0}`, http.StatusBadRequest, jsonResp{false, ""}},
	}

	for _, v := range cases {
		t.Run(v.name, func(t *testing.T) {
			req := httptest.NewRequest("", "/analyze", strings.NewReader(v.req))
			if v.code != http.StatusUnauthorized {
				req.Header.Set("Cookie", cookie)
			}
			w := httptest.NewRecorder()
			analyzeHandler(w, req)

			resp := w.Result()
			body, _ := io.ReadAll(resp.Body)
			if resp.StatusCode != v.code {
				t.Fatalf("want code %v, got code %v, body %v",
					v.code, resp.StatusCode, string(body))
			}

			got := jsonResp{}
			if err := decodeJson(body, &got); err != nil {
				t.Fatalf("fail to decode body: %s", string(body))
			}

			if got.Success != v.want.Success {
				t.Fatalf("want %v, get %v", encodeJson(v.want), string(body))
			}
		})
	}
}
