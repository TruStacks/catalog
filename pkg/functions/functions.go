package functions

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

var dispatcher = newFunctionDispatcher()

type functionDispatcher struct {
	methods map[string]func(map[string]interface{}) (interface{}, error)
}

func newFunctionDispatcher() *functionDispatcher {
	return &functionDispatcher{methods: make(map[string]func(map[string]interface{}) (interface{}, error))}
}

func (fd *functionDispatcher) call(name string, params map[string]interface{}) (interface{}, error) {
	method, ok := fd.methods[name]
	if !ok {
		return nil, errors.New("method not found")
	}
	return method(params)
}

func registerMethod(name string, fn func(map[string]interface{}) (interface{}, error)) {
	dispatcher.methods[name] = fn
}

type rpcCall struct {
	Method string                 `json:"method"`
	Params map[string]interface{} `json:"params"`
}

func FunctionRequestHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err := func() error {
		if err != nil {
			return fmt.Errorf("error reading request body: %s", err)
		}
		var call *rpcCall
		if err := json.Unmarshal(body, &call); err != nil {
			return fmt.Errorf("error parsing json: %s", err)
		}
		result, err := dispatcher.call(call.Method, call.Params)
		if err != nil {
			return err
		}
		data, err := json.Marshal(map[string]interface{}{"result": result})
		if err != nil {
			return fmt.Errorf("error marshalling result: %s", err)
		}
		if _, err := w.Write(data); err != nil {
			return fmt.Errorf("error writing result: %s", err)
		}
		return nil
	}(); err != nil {
		if _, err := w.Write([]byte(fmt.Sprintf(`{"error": "%s"}`, err.Error()))); err != nil {
			log.Println(err)
		}
	}
}
