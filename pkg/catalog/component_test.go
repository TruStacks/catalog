package catalog

import (
	"os"
	"path/filepath"
	"testing"
)

func patchComponentsPath(t *testing.T) func() {
	previousComponentsPath := componentsPath
	d, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	componentsPath = filepath.Join(d, "testdata")
	return func() {
		componentsPath = previousComponentsPath
	}
}

func TestLoadComponentConfig(t *testing.T) {
	defer patchComponentsPath(t)()
	config, err := LoadComponentConfig("test_component")
	if err != nil {
		t.Fatal(err)
	}
	if config.Repo != "https://test.trustacks.charts.io" {
		t.Fatal("configuration was not loaded correctly")
	}
}

func TestDispatcherAddHook(t *testing.T) {
	tests := []struct {
		name     string
		hook     string
		hasError bool
	}{
		{"test", "pre-install", false},
		{"test", "post-install", false},
		{"test", "post-install", true},
	}

	d := newHookDispatcher()
	mockHookFn := func() error { return nil }

	for _, tc := range tests {
		err := d.addHook(tc.name, tc.hook, mockHookFn)
		if !tc.hasError && err != nil {
			t.Fatal("expected an error adding the dispatcher hook")
		} else if tc.hasError && err == nil {
			t.Fatal(err)
		}
	}
}

func TestDispatcherCall(t *testing.T) {
	var x = 0
	increment := func() error {
		x += 1
		return nil
	}
	d := newHookDispatcher()
	if err := d.addHook("test", "increment", increment); err != nil {
		t.Fatal(err)
	}
	if err := d.call("test", "increment"); err != nil {
		t.Fatal(err)
	}
	if x != 1 {
		t.Fatal("got an unexpected value after the dispatch call")
	}
}
