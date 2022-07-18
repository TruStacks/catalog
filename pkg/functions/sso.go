package functions

import (
	"errors"

	"github.com/trustacks/catalog/pkg/components/authentik"
)

var ssoProviderHandlers = map[string]func(params map[string]interface{}) (interface{}, error){
	"authentik": createAuthentikOIDCClient,
}

// CreateOIDCClient creates an openid connection authentication
// client.
func CreateOIDCClient(params map[string]interface{}) (interface{}, error) {
	provider, ok := params["provider"]
	if !ok {
		return nil, errors.New("provider is required")
	}
	method, ok := ssoProviderHandlers[provider.(string)]
	if !ok {
		return nil, errors.New("method handler not foud")
	}
	return method(params)
}

func createAuthentikOIDCClient(params map[string]interface{}) (interface{}, error) {
	p := authentik.CreateOIDCClientParams{
		Name: params["name"].(string),
	}
	return authentik.CreateOIDCClient(p)
}

func init() {
	registerMethod("create-oidc-client", CreateOIDCClient)
}
