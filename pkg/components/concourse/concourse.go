package concourse

import (
	"fmt"
	"log"

	"github.com/trustacks/catalog/pkg/catalog"
)

const componentName = "concourse"

type concourse struct {
	catalog.BaseComponent
}

func (c *concourse) preinstall() error {
	fmt.Println("hello, world!")
	return nil
}

func Include() {
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
	catalog.AddComponent(componentName, component)

	for hook, fn := range map[string]func() error{
		"pre-install": component.preinstall,
	} {
		if err := catalog.AddHook(componentName, hook, fn); err != nil {
			log.Fatal(err)
		}
	}
}
