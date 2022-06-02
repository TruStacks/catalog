package functions

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCallRegisteredMethod(t *testing.T) {
	mockFunction := func(params map[string]interface{}) (interface{}, error) {
		return fmt.Sprintf("hello %s!", params["name"].(string)), nil
	}
	registerMethod("test", mockFunction)
	result, err := Call("test", []byte(`{"name": "world"}`))
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "hello world!", result.(string), "got an unexpected function result")

	_, err = Call("fail", nil)
	assert.Equal(t, err.Error(), "method not found", "expected method not found error")
}
