package flux2

import (
	"github.com/trustacks/catalog/pkg/catalog"
)

type flux2 struct {
	catalog.BaseComponent
}

func init() {
	catalog.AddComponent("flux2", &flux2{
		catalog.BaseComponent{
			Repo:    "https://github.com/fluxcd-community/helm-charts/releases/download/flux2-0.19.2/",
			Chart:   "flux2",
			Version: "0.19.2",
			Hooks:   make([]string, 0),
		},
	})
}
