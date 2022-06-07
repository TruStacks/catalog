package catalog

import (
	"encoding/json"
	"io/ioutil"
	"net/http/httptest"
	"testing"
)

func TestCatalogRequestHandler(t *testing.T) {
	cat := newComponentCatalog()
	cat.addComponent("test", &testComponent{
		&BaseComponent{
			Repo:    "https://charts.test.com",
			Chart:   "test/test",
			Version: "1.0.0",
		},
	})
	w := httptest.NewRecorder()
	catalogRequestHandler(cat)(w, httptest.NewRequest("GET", "https://test.com", nil))
	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)
	comps := make(map[string]interface{})
	if err := json.Unmarshal(body, &comps); err != nil {
		t.Fatal(err)
	}
	if comps["components"].(map[string]interface{})["test"].(map[string]interface{})["repository"].(string) != "https://charts.test.com" {
		t.Fatal("got an unexpected helm repository")
	}
}
