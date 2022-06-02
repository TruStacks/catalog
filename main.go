package main

import "os"

var mode = os.Getenv("CATALOG_MODE")

func main() {
	switch mode {
	default:
		startCatalogServer()
	}
}
