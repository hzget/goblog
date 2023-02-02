package blog

import (
	"fmt"
	"io"
	"net/http"
)

type Handler interface {
	ServeHTTP(http.ResponseWriter, *http.Request)
	Use(http.Handler) Handler
}

type router struct {
	handlers []http.Handler // middleware
	mux      http.Handler
}

func NewHandler() Handler {
	return &router{
		mux: http.DefaultServeMux,
	}
}

func (r *router) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	for _, h := range r.handlers {
		h.ServeHTTP(w, req)
		// TODO:
		//  [1] what if there's some error?
	}

	r.mux.ServeHTTP(w, req)
}

// add a middleware
func (r *router) Use(handler http.Handler) Handler {
	r.handlers = append(r.handlers, handler)
	return r
}

// Middlewares

type httpLogger struct {
	w io.Writer
}

func (l *httpLogger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	headline := fmt.Sprintf("%s %s %s", r.Method, r.URL, r.Proto)
	info := "client request: " + headline
	Debug(info)
	if l.w != nil {
		l.w.Write([]byte(info))
	}
}

func HttpLogger(writer io.Writer) http.Handler {
	return &httpLogger{w: writer}
}
