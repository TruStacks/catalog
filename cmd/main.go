package main

import (
	"log"
	"os"

	"github.com/trustacks/catalog/pkg/catalog"
	"github.com/trustacks/catalog/pkg/components"
	"github.com/trustacks/catalog/pkg/hooks"
	"github.com/trustacks/catalog/server"
)

var (
	mode          = os.Getenv("CATALOG_MODE")
	hookComponent = os.Getenv("HOOK_COMPONENT")
	hookKind      = os.Getenv("HOOK_KIND")
)

func main() {
	cat, err := catalog.NewComponentCatalog()
	if err != nil {
		log.Fatal(err)
	}
	components.Initialize(cat)

	switch mode {
	case "hook":
		if err := hooks.Call(hookComponent, hookKind); err != nil {
			log.Fatal(err)
		}
	default:
		server.StartCatalogServer(cat)
	}
}
