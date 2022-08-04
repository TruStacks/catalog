package catalog

// baseComponent contains default fields and methods for implemented
// components.
type BaseComponent struct {
	Repo             string `json:"repository"`
	Chart            string `json:"chart"`
	Version          string `json:"version"`
	Values           string `json:"values"`
	Hooks            string `json:"hooks"`
	ApplicationHooks string `json:"applicationHooks,omitempty"`
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

// ComponentConfig contains the component configuration fields.
type ComponentConfig struct {
	Repo      string
	Chart     string
	Version   string
	Values    string
	Manifests string
}
