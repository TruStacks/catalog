package functions

import (
	"errors"
	"fmt"

	"github.com/trustacks/catalog/pkg/components/authentik"
)

// CreateOIDCClient creates an openid connection authentication
// client.
func CreateOIDCClient(params map[string]interface{}) (interface{}, error) {
	var method func(string) (interface{}, error)

	provider, ok := params["provider"]
	if !ok {
		return nil, errors.New("provider is required")
	}
	name, ok := params["name"]
	if !ok {
		return nil, errors.New("name is required")
	}

	switch provider.(string) {
	case "authentik":
		method = authentik.CreateOIDCCLient
	default:
		return nil, errors.New("no provider was found to handle the method")
	}
	result, err := method(name.(string))
	if err != nil {
		return nil, fmt.Errorf("error creating the oidc client: %s", err)
	}
	return result, nil
}

func init() {
	registerMethod("create-oidc-client", CreateOIDCClient)
}
