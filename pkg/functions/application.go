package functions

import "errors"

var applicationHandlers = map[string]func(params map[string]interface{}) (interface{}, error){}

// CreateApplication creates an openid connection authentication
// client.
func CreateApplication(params map[string]interface{}) (interface{}, error) {
	provider, ok := params["provider"]
	if !ok {
		return nil, errors.New("provider is required")
	}
	method, ok := applicationHandlers[provider.(string)]
	if !ok {
		return nil, errors.New("method handler not foud")
	}
	return method(params)
}

func init() {
	registerMethod("create-application", CreateApplication)
}
