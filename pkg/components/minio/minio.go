package minio

import (
	"github.com/trustacks/catalog/pkg/catalog"
)

type minio struct {
	catalog.BaseComponent
}

func init() {
	catalog.AddComponent("minio", &minio{
		catalog.BaseComponent{
			Repo:    "https://charts.min.io/helm-releases",
			Chart:   "minio",
			Version: "4.0.2",
			Hooks:   make([]string, 0),
		},
	})
}
