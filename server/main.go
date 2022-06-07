package main

import (
	"fmt"
	"os"

	"github.com/trustacks/catalog/pkg/catalog"

	_ "github.com/trustacks/catalog/pkg/components"
)

var mode = os.Getenv("CATALOG_MODE")

func main() {
	switch mode {
	case "hook":
		fmt.Println("hook mode")
	default:
		catalog.StartCatalogServer()
	}
}
