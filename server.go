package main

import (
	"encoding/json"
	"net/http"
	"os"
)

// Server is our webserver state
type Server struct {
	c Config
}

// NewServer creates a new Server
func NewServer(c Config) Server {
	return Server{c: c}
}

func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	v, err := DecryptURL(s.c, r.URL)
	if err != nil {
		w.WriteHeader(404)
		return
	}

	dest, ok := v["url"]
	if !ok {
		w.WriteHeader(404)
		return
	}

	destinationURL, ok := dest.(string)
	if !ok {
		w.WriteHeader(404)
		return
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(v)
	http.Redirect(w, r, destinationURL, 301)
}
