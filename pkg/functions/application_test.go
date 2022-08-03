package functions

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateApplication(t *testing.T) {
	createApplicationHandler["test"] = func(params map[string]interface{}) (interface{}, error) {
		return 42, nil
	}
	result, err := Call("create-application", []byte(`{"provider": "test"}`))
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 42, result.(int), "got an unexpected result")
}
