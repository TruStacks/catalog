package functions

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSSOHandler(t *testing.T) {
	ssoProviderHandlers["test"] = func(params map[string]interface{}) (interface{}, error) {
		return 42, nil
	}
	result, err := Call("create-oidc-client", map[string]interface{}{"provider": "test"})
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 42, result.(int), "got an unexpected result")
}
