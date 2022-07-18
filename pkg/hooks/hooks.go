package hooks

import "fmt"

// errHookAlreadyExists is returned if the hook already exists for
// the component.
var errHookAlreadyExists = fmt.Errorf("the hook already exists")

// hookDispatcher is used to call hooks.
var hookDispatcher = newHookDispatcher()

// dispatcher manages calls to component hooks.
type dispatcher struct {
	methods map[string]map[string]func() error
}

// hook names.
const (
	PreInstallHook  = "pre-install"
	PostInstallHook = "post-install"
	PreDeleteHook   = "pre-delete"
	PostDeleteHook  = "post-delete"
	PreUpgrade      = "pre-upgrade"
	PostUpgrade     = "post-upgrade"
	PreRollback     = "pre-rollback"
	PostRollback    = "post-rollback"
)

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
func Call(component, hook string) error {
	return hookDispatcher.call(component, hook)
}
