package vault

import (
	"log"

	"github.com/trustacks/catalog/pkg/catalog"
)

const componentName = "vault"

type vault struct {
	catalog.BaseComponent
}

// Initialize adds the component to the catalog and configures hooks.
func Initialize() {
	config, err := catalog.LoadComponentConfig(componentName)
	if err != nil {
		log.Fatal(err)
	}
	component := &vault{
		catalog.BaseComponent{
			Repo:       config.Repo,
			Chart:      config.Chart,
			Version:    config.Version,
			Values:     config.Values,
			Hooks:      config.Hooks,
			Parameters: config.Parameters,
		},
	}
	catalog.AddComponent(componentName, component)
}
