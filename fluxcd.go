package main

type fluxcd struct {
	baseComponent
}

func init() {
	catalog.addComponent("fluxcd", &fluxcd{
		baseComponent{
			Repo:    "https://fluxcd-community.github.io/helm-charts",
			Chart:   "fluxcd-community/flux2",
			Version: "0.19.2",
			Hooks:   make([]string, 0),
		},
	})
}
