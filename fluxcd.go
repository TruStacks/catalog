package main

type fluxcd struct {
	baseComponent
}

func init() {
	catalog.addComponent("fluxcd", &fluxcd{
		baseComponent{
			Repo:    "https://github.com/fluxcd-community/helm-charts/releases/download/flux2-0.19.2/",
			Chart:   "flux2",
			Version: "0.19.2",
			Hooks:   make([]string, 0),
		},
	})
}
