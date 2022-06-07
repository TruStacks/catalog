package main

type minio struct {
	baseComponent
}

func init() {
	catalog.addComponent("minio", &minio{
		baseComponent{
			Repo:    "https://charts.min.io/helm-releases",
			Chart:   "minio",
			Version: "4.0.2",
			Hooks:   make([]string, 0),
		},
	})
}
