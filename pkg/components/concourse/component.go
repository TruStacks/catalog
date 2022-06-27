package concourse

import (
	"log"

	"github.com/trustacks/catalog/pkg/catalog"
)

const componentName = "concourse"

type concourse struct {
	catalog.BaseComponent
}

// Initialize adds the component to the catalog and configures hooks.
func Initialize(c *catalog.ComponentCatalog) {
	config, err := catalog.LoadComponentConfig(componentName)
	if err != nil {
		log.Fatal(err)
	}
	component := &concourse{
		catalog.BaseComponent{
			Repo:    config.Repo,
			Chart:   config.Chart,
			Version: config.Version,
			Values:  config.Values,
			Hooks:   config.Hooks,
		},
	}
	c.AddComponent(componentName, component)
}
