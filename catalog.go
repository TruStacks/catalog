package main

import (
	"encoding/json"
	"log"
	"net/http"
)

var catalog = newComponentCatalog()

type component interface {
	repo() string
	chart() string
	preInstall() error
	postInstall() error
	preDelete() error
	postDelete() error
	preUpgrade() error
	postUpgrade() error
	preRollback() error
	postRollback() error
}

type componentCatalog struct {
	Components map[string]component `json:"components"`
}

func (c *componentCatalog) addComponent(name string, component component) {
	c.Components[name] = component
}

func newComponentCatalog() *componentCatalog {
	return &componentCatalog{make(map[string]component)}
}

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

func startCatalogServer() {
	http.HandleFunc("/.well-known/catalog-manifest", catalogRequestHandler(catalog))
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
