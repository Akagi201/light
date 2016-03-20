package main

import (
	"encoding/json"
	"net/http"

	"github.com/Akagi201/light/api"
)

type Message struct {
	Text string
}

func handleHello(w http.ResponseWriter, r *http.Request) string {
	m := Message{"Hello World"}
	b, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}

	return string(b)
}

func main() {
	handlers := map[string]func(http.ResponseWriter, *http.Request){}
	handlers["/hello/"] = func(w http.ResponseWriter, r *http.Request) {
		api.Respond(map[string]string{"foo": "bar"}, handleHello)(w, r)
	}
	c := api.New("localhost", 1024, 10, handlers)
	c.Run()
}
