package functions

// PatchMockFunction patches the dispatcher with the mock function.
func PatchMockFunction(name string, fn func(params map[string]interface{}) (interface{}, error)) func() {
	previousMethod := dispatcher.methods[name]
	dispatcher.methods[name] = fn
	return func() {
		dispatcher.methods[name] = previousMethod
	}
}
