package main

import (
	"os"

	"github.com/trustacks/catalog/pkg/catalog"

	_ "github.com/trustacks/catalog/pkg/components"
)

var (
	mode      = os.Getenv("CATALOG_MODE")
	component = os.Getenv("HOOK_COMPONENT")
	hookKind  = os.Getenv("HOOK_KIND")
)

func main() {
	switch mode {
	case "hook":
		catalog.Call(component, hookKind)
	default:
		catalog.StartCatalogServer()
	}
}
