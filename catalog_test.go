package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http/httptest"
	"os"
	"testing"
)

type testComponent struct {
	*baseComponent
}

func TestCatalogAddComponent(t *testing.T) {
	cat := newComponentCatalog()
	cat.addComponent("test", &testComponent{
		&baseComponent{
			Repo:    "https://charts.test.com",
			Chart:   "test/test",
			Version: "1.0.0",
		},
	})
	if cat.Components["test"].repo() != "https://charts.test.com" {
		t.Fatal("got an unexpected helm repository")
	}
}

func TestCatalogRequestHandler(t *testing.T) {
	cat := newComponentCatalog()
	cat.addComponent("test", &testComponent{
		&baseComponent{
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

func TestCatalogHookSource(t *testing.T) {
	defer func() {
		catalogHookSource = os.Getenv("CATALOG_HOOK_SOURCE")
	}()
	catalogHookSource = "test-registry/trustacks/hooks"
	w := httptest.NewRecorder()
	catalogRequestHandler(newComponentCatalog())(w, httptest.NewRequest("GET", "https://test.com", nil))
	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)
	comps := make(map[string]interface{})
	if err := json.Unmarshal(body, &comps); err != nil {
		t.Fatal(err)
	}
	if comps["hookSource"].(string) != "test-registry/trustacks/hooks" {
		t.Fatal("got an unexpected hook source")
	}
}
