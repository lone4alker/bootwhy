package rules

import "testing"

func TestDefaultRulesAreActionable(t *testing.T) {
	for _, hypothesis := range Default() {
		if hypothesis.Source == "" || hypothesis.Risk == "" {
			t.Fatalf("incomplete hypothesis %#v", hypothesis)
		}
		if hypothesis.Option != "" && hypothesis.Undo == "" {
			t.Fatalf("option without undo %#v", hypothesis)
		}
	}
}

func TestDeviceAndNetworkRules(t *testing.T) {
	for _, name := range []string{"device-timeout", "network-online-wait"} {
		if !hasRule(name) {
			t.Fatalf("missing %s", name)
		}
	}
}

func TestFsckFailureAndUdevRules(t *testing.T) {
	for _, name := range []string{"fsck", "failing-unit", "udev-settle"} {
		if !hasRule(name) {
			t.Fatalf("missing %s", name)
		}
	}
}

func hasRule(name string) bool {
	for _, hypothesis := range Default() {
		if hypothesis.Name == name {
			return true
		}
	}
	return false
}
