package concourse

import (
	"log"

	"github.com/trustacks/catalog/pkg/catalog"
)

const componentName = "concourse"

type concourse struct {
	catalog.BaseComponent
}

func (c *concourse) preinstall() error {
	return nil
}

func Include() {
	config, err := catalog.LoadComponentConfig(componentName)
	if err != nil {
		log.Fatal(err)
	}
	catalog.AddComponent(componentName, &concourse{
		catalog.BaseComponent{
			Repo:    config.Repo,
			Chart:   config.Chart,
			Version: config.Version,
			Values:  config.Values,
			Hooks:   config.Hooks,
		},
	})
}
