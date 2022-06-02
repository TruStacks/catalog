package server

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/trustacks/catalog/pkg/catalog"
)

type testComponent struct {
	*catalog.BaseComponent
}

func TestCatalogRequestHandler(t *testing.T) {
	cat, err := catalog.NewComponentCatalog()
	if err != nil {
		t.Fatal(err)
	}
	cat.AddComponent("test", &testComponent{
		&catalog.BaseComponent{
			Repo:    "https://charts.test.com",
			Chart:   "test/test",
			Version: "1.0.0",
		},
	})
	w := httptest.NewRecorder()
	catalogRequestHandler(cat)(w, httptest.NewRequest("GET", "https://test.com", nil))
	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	comps := make(map[string]interface{})
	if err := json.Unmarshal(body, &comps); err != nil {
		t.Fatal(err)
	}
	if comps["components"].(map[string]interface{})["test"].(map[string]interface{})["repository"].(string) != "https://charts.test.com" {
		t.Fatal("got an unexpected helm repository")
	}
}
