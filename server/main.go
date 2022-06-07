package main

import (
	"os"

	"github.com/trustacks/catalog/pkg/catalog"

	_ "github.com/trustacks/catalog/pkg/components"
)

var mode = os.Getenv("CATALOG_MODE")

func main() {
	switch mode {
	default:
		catalog.StartCatalogServer()
	}
}
