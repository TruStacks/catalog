package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

// catalogHookSource is the source for the hook container.
var catalogHookSource = os.Getenv("CATALOG_HOOK_SOURCE")

// catalog is a singleton for components to register themselves
// in the catalog manifest.
var catalog = newComponentCatalog()

// componentCatalog contains the component manifests.
type componentCatalog struct {
	HookSource string               `json:"hookSource"`
	Components map[string]component `json:"components"`
}

// addComponent adds the component to the catalog.
func (c *componentCatalog) addComponent(name string, component component) {
	c.Components[name] = component
}

// newComponentCatalog creates the
func newComponentCatalog() *componentCatalog {
	return &componentCatalog{catalogHookSource, make(map[string]component)}
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

		if _, err := w.Write(data); err != nil {
			log.Println("error:", err)
		}
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
