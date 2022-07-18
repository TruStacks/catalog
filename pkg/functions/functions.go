package functions

import (
	"errors"
)

// dispatcher is the global function dispatcher.
var dispatcher = newFunctionDispatcher()

// functionDispatcher contains methods used for intercomponent
// tasks.
type functionDispatcher struct {
	methods map[string]func(map[string]interface{}) (interface{}, error)
}

// newFunctionDispatcher creates a function dispatcher instance.
func newFunctionDispatcher() *functionDispatcher {
	return &functionDispatcher{methods: make(map[string]func(map[string]interface{}) (interface{}, error))}
}

// call executes the target method with the provided function
// parameters.
func (fd *functionDispatcher) call(name string, params map[string]interface{}) (interface{}, error) {
	method, ok := fd.methods[name]
	if !ok {
		return nil, errors.New("method not found")
	}
	return method(params)
}

// registerMethod add the method to the function dispatcher.
func registerMethod(name string, fn func(map[string]interface{}) (interface{}, error)) {
	dispatcher.methods[name] = fn
}

// Call sends the method parameters the function dispatcher for
// execution.
func Call(name string, params map[string]interface{}) (interface{}, error) {
	return dispatcher.call(name, params)
}
