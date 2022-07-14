package argocd

import (
	_ "embed"
	"log"

	"github.com/trustacks/catalog/pkg/catalog"
	"gopkg.in/yaml.v2"
)

const componentName = "argocd"

type argocd struct {
	catalog.BaseComponent
}

//go:embed config.yaml
var config []byte

// Initialize adds the component to the catalog and configures hooks.
func Initialize(c *catalog.ComponentCatalog) {
	var conf *catalog.ComponentConfig
	if err := yaml.Unmarshal(config, &conf); err != nil {
		log.Fatal(err)
	}
	component := &argocd{
		catalog.BaseComponent{
			Repo:    conf.Repo,
			Chart:   conf.Chart,
			Version: conf.Version,
			Values:  conf.Values,
			Hooks:   conf.Hooks,
		},
	}
	c.AddComponent(componentName, component)

	for hook, fn := range map[string]func() error{} {
		if err := catalog.AddHook(componentName, hook, fn); err != nil {
			log.Fatal(err)
		}
	}
}
