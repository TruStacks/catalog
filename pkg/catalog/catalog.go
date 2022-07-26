package catalog

import (
	_ "embed"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

var (
	// catalogHookSource is the source for the hook container.
	catalogHookSource = fmt.Sprintf("%s:%s", os.Getenv("CATALOG_HOOK_SOURCE"), os.Getenv("CATALOG_VERSION"))
)

// component contains methods for components when running in hook
// mode.
type component interface {
	repo() string
	chart() string
	version() string
	preInstall() error
	postInstall() error
	preDelete() error
	postDelete() error
	preUpgrade() error
	postUpgrade() error
	preRollback() error
	postRollback() error
}

type componentCatalogConfigParameters struct {
	Name    string `json:"name"`
	Default string `json:"default"`
}

type componentCatalogConfig struct {
	Parameters []componentCatalogConfigParameters `json:"parameters"`
}

// ComponentCatalog contains the component manifests.
type ComponentCatalog struct {
	HookSource string                  `json:"hookSource"`
	Components map[string]component    `json:"components"`
	Config     *componentCatalogConfig `json:"config"`
}

// addComponent adds the component to the catalog.
func (c *ComponentCatalog) AddComponent(name string, component component) {
	c.Components[name] = component
}

// loadConfig loads the catalog configuration yaml file.
func loadConfig(data []byte) (*componentCatalogConfig, error) {
	var config *componentCatalogConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return config, nil
}

//go:embed catalog.yaml
var catalogConfig []byte

// newComponentCatalog creates the
func NewComponentCatalog() (*ComponentCatalog, error) {
	config, err := loadConfig(catalogConfig)
	if err != nil {
		return nil, err
	}
	return &ComponentCatalog{
		HookSource: catalogHookSource,
		Components: make(map[string]component),
		Config:     config,
	}, nil
}
