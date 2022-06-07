package catalog

import (
	"os"
)

// catalogHookSource is the source for the hook container.
var catalogHookSource = os.Getenv("CATALOG_HOOK_SOURCE")

// catalog is a singleton for components to register themselves
// in the catalog manifest.
var catalog = newComponentCatalog()

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

// baseComponent contains default fields and methods for implemented
// components.
type BaseComponent struct {
	Repo    string   `json:"repository"`
	Chart   string   `json:"chart"`
	Version string   `json:"version"`
	Values  string   `json:"values"`
	Hooks   []string `json:"hooks"`
}

// repo returns the component's helm repository.
func (c *BaseComponent) repo() string {
	return c.Repo
}

// chart returns the component's helm chart.
func (c *BaseComponent) chart() string {
	return c.Chart
}

// version returns the component's helm chart version.
func (c *BaseComponent) version() string {
	return c.Version
}

// preInstall executes after templates are rendered, but before any
// resources are created in kubernetes.
func (c *BaseComponent) preInstall() error {
	return nil
}

// postInstall executes after all resources are loaded into
// kubernetes.
func (c *BaseComponent) postInstall() error {
	return nil
}

// preDelete executes on a deletion request before any resources are
// deleted from kubernetes.
func (c *BaseComponent) preDelete() error {
	return nil
}

// postDelete executes on a deletion request after all of the
// release's resources have been deleted.
func (c *BaseComponent) postDelete() error {
	return nil
}

// preUpgrade executes on an upgrade request after templates are
// rendered, but before any resources are updated.
func (c *BaseComponent) preUpgrade() error {
	return nil
}

// postUpgrade executes on an upgrade request after all resources
// have been upgraded.
func (c *BaseComponent) postUpgrade() error {
	return nil
}

// preRollback executes on a rollback request after templates are
// rendered, but before any resources are rolled back.
func (c *BaseComponent) preRollback() error {
	return nil
}

// postRollback executes on a rollback request after all resources
// have been modified.
func (c *BaseComponent) postRollback() error {
	return nil
}

// componentCatalog contains the component manifests.
type componentCatalog struct {
	HookSource string               `json:"hookSource"`
	Components map[string]component `json:"components"`
}

// addComponent adds the component to the catalog.
func (c *componentCatalog) addComponent(name string, component component) {
	c.Components[name] = component
}

// AddComponent adds the component to the catalog singleton.
func AddComponent(name string, component component) {
	catalog.addComponent(name, component)
}

// newComponentCatalog creates the
func newComponentCatalog() *componentCatalog {
	return &componentCatalog{catalogHookSource, make(map[string]component)}
}
