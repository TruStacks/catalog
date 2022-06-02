package main

import (
	"encoding/json"
	"log"
	"net/http"
)

// catalog is a singleton for components to register themselves
// in the catalog manifest.
var catalog = newComponentCatalog()

// componentCatalog contains the component manifests.
type componentCatalog struct {
	Components map[string]component `json:"components"`
}

// addComponent adds the component to the catalog.
func (c *componentCatalog) addComponent(name string, component component) {
	c.Components[name] = component
}

// newComponentCatalog creates the
func newComponentCatalog() *componentCatalog {
	return &componentCatalog{make(map[string]component)}
}

// catalogRequestHandler returns the component catalog json
// manifest.
func catalogRequestHandler(c *componentCatalog) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := json.Marshal(c)
		if err != nil {
			log.Println("error unmarshaling the catalog:", err)
		}
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}
}

// startCatalogServer starts the catalog server.
func startCatalogServer() {
	http.HandleFunc("/.well-known/catalog-manifest", catalogRequestHandler(catalog))
	log.Println("starting server on *:8080	")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
