package catalog

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// componentsPath is the path to the components sources.
var componentsPath = "/components"

// componentConfig contains the component configuration fields.
type componentConfig struct {
	Repo       string
	Chart      string
	Version    string
	Values     string
	Hooks      []string
	Parameters []map[string]interface{}
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

// dispatcher is used to call hooks.
var dispatcher = newHookDispatcher()

// hookDispatcher manages calls to component hooks.
type hookDispatcher struct {
	hooks map[string]map[string]func() error
}

// AddHook adds the component hook to the disptacher.
func (d *hookDispatcher) addHook(component, hook string, fn func() error) error {
	if _, ok := d.hooks[component]; ok {
		if _, ok := d.hooks[component][hook]; ok {
			return errHookAlreadyExists
		}
		d.hooks[component][hook] = fn
	} else {
		d.hooks[component] = map[string]func() error{hook: fn}
	}

	return nil
}

// Call executes the component hook.
func (d *hookDispatcher) call(component, hook string) error {
	return d.hooks[component][hook]()
}

// newHookDispatcher creates a new hook dispatcher instance
func newHookDispatcher() *hookDispatcher {
	return &hookDispatcher{make(map[string]map[string]func() error)}
}

// AddHook adds the hook to the global dispatcher.
func AddHook(component, hook string, fn func() error) error {
	return dispatcher.addHook(component, hook, fn)
}

// Call runs the hook using the global dispatcher.
func Call(component, hook string) error {
	return dispatcher.call(component, hook)
}
