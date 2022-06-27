package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/trustacks/catalog/pkg/catalog"

	// initialize the components.
	"github.com/trustacks/catalog/pkg/components"
)

// serverPort is the port of the webserver.
const serverPort = "80"

// catalogRequestHandler returns the component catalog json
// manifest.
func catalogRequestHandler(c *catalog.ComponentCatalog) func(w http.ResponseWriter, r *http.Request) {
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
func StartCatalogServer() {
	cat, err := catalog.NewComponentCatalog()
	if err != nil {
		log.Fatal(err)
	}
	components.Initialize(cat)
	http.HandleFunc("/.well-known/catalog-manifest", catalogRequestHandler(cat))
	log.Printf("starting server on *:%s\n", serverPort)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", serverPort), nil); err != nil {
		log.Fatal(err)
	}
}
