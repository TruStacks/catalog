package vault

import (
	"github.com/trustacks/catalog/pkg/catalog"
)

type vault struct {
	catalog.BaseComponent
}

func init() {
	catalog.AddComponent("vault", &vault{
		catalog.BaseComponent{
			Repo:    "https://helm.releases.hashicorp.com",
			Chart:   "vault",
			Version: "0.20.1",
			Hooks:   make([]string, 0),
		},
	})
}
