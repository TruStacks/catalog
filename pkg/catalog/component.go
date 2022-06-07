package catalog

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// componentsPath is the path to the components sources.
var componentsPath string

// componentConfig contains the component configuration fields.
type componentConfig struct {
	Repo    string                 `json:"repository"`
	Chart   string                 `json:"chart"`
	Version string                 `json:"version"`
	Values  map[string]interface{} `json:"values"`
	Hooks   []string               `json:"hooks"`
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

// dispatcher manages calls to component hooks.
type dispatcher struct {
	hooks map[string]map[string]func() error
}

// AddHook adds the component hook to the disptacher.
func (d *dispatcher) AddHook(component, hook string, fn func() error) error {
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
func (d *dispatcher) Call(component, hook string) error {
	return d.hooks[component][hook]()
}

// newDispatcher creates a new dispatcher instance
func newDispatcher() *dispatcher {
	return &dispatcher{make(map[string]map[string]func() error)}
}

func init() {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	componentsPath = filepath.Join(dir, "pkg", "components")
}
