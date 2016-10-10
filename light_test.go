package light_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/Akagi201/light"
	"github.com/ivpusic/httpcheck"
)

// TestGet test GET
func TestGet(t *testing.T) {
	app := light.New()

	app.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello"))
	})

	checker := httpcheck.New(t, app)

	checker.Test("GET", "/").
		Check().
		HasStatus(200).
		HasBody([]byte("hello"))
}

// TestHead test HEAD
func TestHead(t *testing.T) {
	app := light.New()

	app.Head("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello"))
	})

	checker := httpcheck.New(t, app)

	checker.Test("HEAD", "/").
		Check().
		HasStatus(200)
}

// TestHeadGet test HEAD for GET route
func TestHeadGet(t *testing.T) {
	app := light.New()

	app.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello"))
	})

	checker := httpcheck.New(t, app)

	checker.Test("HEAD", "/").
		Check().
		HasStatus(200)
}

// TestPrecedence test route precedence
func TestPrecedence(t *testing.T) {
	app := light.New()

	app.Get("/foo", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello"))
	})

	app.Get("/foo", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("world"))
	})

	checker := httpcheck.New(t, app)

	checker.Test("GET", "/foo").
		Check().
		HasStatus(200).
		HasBody([]byte("hello"))
}

// TestMany test many routes
func TestMany(t *testing.T) {
	app := light.New()

	app.Get("/foo", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello"))
	})

	app.Get("/bar", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("world"))
	})

	checker := httpcheck.New(t, app)

	checker.Test("GET", "/foo").
		Check().
		HasStatus(200).
		HasBody([]byte("hello"))

	checker.Test("GET", "/bar").
		Check().
		HasStatus(200).
		HasBody([]byte("world"))
}

// TestParams test params
func TestParams(t *testing.T) {
	app := light.New()

	app.Get("/user/:name/pet/:pet", func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get(":name")
		pet := r.URL.Query().Get(":pet")
		fmt.Fprintf(w, "user %s's pet %s", name, pet)
	})

	checker := httpcheck.New(t, app)

	checker.Test("GET", "/user/tobi/pet/loki").
		Check().
		HasStatus(200).
		HasBody([]byte("user tobi's pet loki"))
}
