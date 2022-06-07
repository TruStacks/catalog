package concourse

import (
	"github.com/trustacks/catalog/pkg/catalog"
)

type concourse struct {
	catalog.BaseComponent
}

func init() {
	catalog.AddComponent("concourse", &concourse{
		catalog.BaseComponent{
			Repo:    "https://concourse-charts.storage.googleapis.com",
			Chart:   "concourse",
			Version: "16.1.22",
			Hooks:   make([]string, 0),
		},
	})
}
