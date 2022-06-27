package catalog

import (
	"io/ioutil"
	"os"
	"testing"
)

type testComponent struct {
	*BaseComponent
}

func TestCatalogAddComponent(t *testing.T) {
	cat, err := NewComponentCatalog()
	if err != nil {
		t.Fatal(err)
	}
	cat.AddComponent("test", &testComponent{
		&BaseComponent{
			Repo:    "https://charts.test.com",
			Chart:   "test/test",
			Version: "1.0.0",
		},
	})
	if cat.Components["test"].repo() != "https://charts.test.com" {
		t.Fatal("got an unexpected helm repository")
	}
}

func TestCatalogLoadConfig(t *testing.T) {
	f, err := ioutil.TempFile("", "config")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write([]byte(`parameters:
- name: test
  default: default
    `)); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	config, err := loadConfig(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	if config.Parameters[0].Name != "test" {
		t.Fatal("got an unexpected parameter name")
	}
	if config.Parameters[0].Default != "default" {
		t.Fatal("got an unexpected parameter default value")
	}
}

func TestCatalogHookSource(t *testing.T) {
	defer func() {
		catalogHookSource = os.Getenv("CATALOG_HOOK_SOURCE")
	}()
	catalogHookSource = "test-registry/trustacks/hooks"
	cat, err := NewComponentCatalog()
	if err != nil {
		t.Fatal(err)
	}
	if cat.HookSource != "test-registry/trustacks/hooks" {
		t.Fatal("got an unexpected hook source")
	}
}
