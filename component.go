package main

type baseComponent struct {
	Repo    string `json:"repository"`
	Chart   string `json:"chart"`
	Version string `json:"version"`
}

// repo returns the component's helm repository.
func (c *baseComponent) repo() string {
	return c.Repo
}

// chart returns the component's helm chart.
func (c *baseComponent) chart() string {
	return c.Chart
}

// version returns the component's helm chart version.
func (c *baseComponent) version() string {
	return c.Version
}

// preInstall executes after templates are rendered, but before any
// resources are created in kubernetes.
func (c *baseComponent) preInstall() error {
	return nil
}

// postInstall executes after all resources are loaded into
// kubernetes.
func (c *baseComponent) postInstall() error {
	return nil
}

// preDelete executes on a deletion request before any resources are
// deleted from kubernetes.
func (c *baseComponent) preDelete() error {
	return nil
}

// postDelete executes on a deletion request after all of the
// release's resources have been deleted.
func (c *baseComponent) postDelete() error {
	return nil
}

// preUpgrade executes on an upgrade request after templates are
// rendered, but before any resources are updated.
func (c *baseComponent) preUpgrade() error {
	return nil
}

// postUpgrade executes on an upgrade request after all resources
// have been upgraded.
func (c *baseComponent) postUpgrade() error {
	return nil
}

// preRollback executes on a rollback request after templates are
// rendered, but before any resources are rolled back.
func (c *baseComponent) preRollback() error {
	return nil
}

// postRollback executes on a rollback request after all resources
// have been modified.
func (c *baseComponent) postRollback() error {
	return nil
}
