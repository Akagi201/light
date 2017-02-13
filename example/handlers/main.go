package main

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/Akagi201/light"
	"github.com/gohttp/logger"
)

func authHandlerFunc(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()
		if ok && u == "foobar" && p == "foobared" {
			ctx := light.Context(r)
			ctx = context.WithValue(ctx, "user", u)
			light.SetContext(ctx, r)
			next(w, r)
			return
		}
		w.Header().Set("WWW-Authenticate", `realm="restricted"`)
		w.WriteHeader(http.StatusUnauthorized)
	}
}

func authHandler(next http.Handler) http.Handler {
	return authHandlerFunc(next.ServeHTTP)
}

func main() {
	// curl -i localhost:8080
	// curl -i -XPOST --basic -u foobar:foobared localhost:8080/auth/login
	root := light.New()
	root.Use(logger.New())
	root.Get("/", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "hello, world\n")
	})
	auth := light.New()
	{
		auth.Use(authHandler)
		auth.Post("/login", func(w http.ResponseWriter, r *http.Request) {
			u := light.Context(r).Value("user")
			fmt.Fprintln(w, "hello,", u)
		})
	}
	root.Append("/auth", auth)
	http.ListenAndServe(":8080", root)
}
