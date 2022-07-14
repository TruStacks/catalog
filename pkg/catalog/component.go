package catalog

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

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

// componentConfig contains the component configuration fields.
type componentConfig struct {
	Repo    string
	Chart   string
	Version string
	Values  string
	Hooks   []string
}

// LoadComponentConfig loads the component configuration values from
// the config file.
func LoadComponentConfig(component string) (*componentConfig, error) {
	data, err := ioutil.ReadFile(filepath.Join(componentsPath, component, "config.yaml"))
	if err != nil {
		return nil, err
	}
	var config *componentConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return config, nil
}

// errHookAlreadyExists is returned if the hook already exists for
// the component.
var errHookAlreadyExists = fmt.Errorf("the hook already exists")

// hookDispatcher is used to call hooks.
var hookDispatcher = newHookDispatcher()

// dispatcher manages calls to component hooks.
type dispatcher struct {
	methods map[string]map[string]func() error
}

// AddHook adds the component hook to the disptacher.
func (d *dispatcher) addHook(component, hook string, fn func() error) error {
	if _, ok := d.methods[component]; ok {
		if _, ok := d.methods[component][hook]; ok {
			return errHookAlreadyExists
		}
		d.methods[component][hook] = fn
	} else {
		d.methods[component] = map[string]func() error{hook: fn}
	}

	return nil
}

// Call executes the component hook.
func (d *dispatcher) call(component, hook string) error {

	return d.methods[component][hook]()
}

// newHookDispatcher creates a new hook dispatcher instance
func newHookDispatcher() *dispatcher {
	return &dispatcher{make(map[string]map[string]func() error)}
}

// AddHook adds the hook to the global dispatcher.
func AddHook(component, hook string, fn func() error) error {
	return hookDispatcher.addHook(component, hook, fn)
}

// Call runs the hook using the global dispatcher.
func CallHook(component, hook string) error {
	return hookDispatcher.call(component, hook)
}
