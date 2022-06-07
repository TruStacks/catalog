package catalog

import (
	"encoding/json"
	"log"
	"net/http"
)

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
func StartCatalogServer() {
	http.HandleFunc("/.well-known/catalog-manifest", catalogRequestHandler(catalog))
	log.Println("starting server on *:8080	")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
