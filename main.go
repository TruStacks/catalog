package main

import "os"

var mode = os.Getenv("MODE")

func main() {
	switch mode {
	case "server":
		startCatalogServer()
	}
}
