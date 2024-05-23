package server

import (
	"encoding/json"
	_ "github.com/99designs/gqlgen/graphql/introspection"
	"log"
	"net/http"
)

func GetIndexHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(struct {
			Name    string
			Version string
		}{
			Name:    "Memo GraphQL",
			Version: "0.1",
		}); err != nil {
			log.Printf("error marshalling and writing memo graph version; %v", err)
		}
		log.Printf("Processed graph request: %s\n", r.URL)
	}
}
