package functions

import (
	"errors"
)

var createOIDCclientHandlers = map[string]func(params map[string]interface{}) (interface{}, error){}

// createOIDCClient creates an openid connection authentication
// client.
func createOIDCClient(params map[string]interface{}) (interface{}, error) {
	provider, ok := params["provider"]
	if !ok {
		return nil, errors.New("provider is required")
	}
	method, ok := createOIDCclientHandlers[provider.(string)]
	if !ok {
		return nil, errors.New("method handler not foud")
	}
	return method(params)
}

// AddCreateOIDCClientHandler adds the create oidc client handler
// method.
func AddCreateOIDCClientHandler(name string, handler func(params map[string]interface{}) (interface{}, error)) {
	createOIDCclientHandlers[name] = handler
}

func init() {
	registerMethod("create-oidc-client", createOIDCClient)
}
