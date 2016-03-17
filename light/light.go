package light

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type Context struct {
	mux    *http.ServeMux
	server *http.Server
	Logger *log.Logger
}

func Respond(headers map[string]interface{}, fn func(w http.ResponseWriter, r *http.Request) string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		for k, v := range headers {
			w.Header().Set(k, v.(string))
		}
		data := fn(w, r)
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))
		fmt.Fprintf(w, data)
	}
}

func New(host string, port, timeout int, handlers map[string]func(http.ResponseWriter, *http.Request)) (context *Context) {
	mux := http.NewServeMux()
	for pattern, handler := range handlers {
		mux.Handle(pattern, http.HandlerFunc(handler))
	}

	s := &http.Server{
		Addr:        fmt.Sprintf("%s:%d", host, port),
		Handler:     mux,
		ReadTimeout: time.Duration(timeout) * time.Second, // to prevent abuse of "keep-alive" requests by clients
	}

	context = &Context{
		mux:    mux,
		server: s,
		Logger: log.New(os.Stdout, "", log.Ldate|log.Ltime),
	}

	return
}

func (context *Context) Run() {
	// serve requests using the default http.Server
	context.server.ListenAndServe()
}
