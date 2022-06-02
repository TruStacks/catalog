package main

type vault struct {
	baseComponent
}

func init() {
	catalog.addComponent("vault", &vault{
		baseComponent{
			Repo:    "https://helm.releases.hashicorp.com",
			Chart:   "hashicorp/vault",
			Version: "0.20.1",
			Hooks:   make([]string, 0),
		},
	})
}
