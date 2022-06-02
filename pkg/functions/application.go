package functions

import (
	"errors"
)

var createApplicationHandler = make(map[string]func(params map[string]interface{}) (interface{}, error))

// CreateApplication creates an openid connection authentication
// client.
func CreateApplication(params map[string]interface{}) (interface{}, error) {
	provider, ok := params["provider"]
	if !ok {
		return nil, errors.New("provider is required")
	}
	method, ok := createApplicationHandler[provider.(string)]
	if !ok {
		return nil, errors.New("method handler not foud")
	}
	return method(params)
}

// AddCreateApplicationHandler adds the create application handler
// method.
func AddCreateApplicationHandler(name string, handler func(params map[string]interface{}) (interface{}, error)) {
	createApplicationHandler[name] = handler
}

func init() {
	registerMethod("create-application", CreateApplication)
}