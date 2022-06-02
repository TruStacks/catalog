package hooks

import "testing"

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
