package main

import (
	"fmt"
	"net/http"

	"github.com/Akagi201/light"
)

func main() {
	app := light.New()

	app.Get("/foo", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "foo")
	}))

	app.Get("/bar", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "bar")
	})

	app.Get("/user/:name/pet/:pet", func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get(":name")
		pet := r.URL.Query().Get(":pet")
		fmt.Fprintf(w, "user %s's pet %s", name, pet)
	})

	app.Listen(":3000")
}
