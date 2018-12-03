package handlers

import (
	"testing"

	"github.com/flavio/collage/config"
)

func TestTranslateName(t *testing.T) {
	configData := `
	{
		"mappings" : {
			"cool/stuff" : "index.docker.io/flavio",
			"cool/distro" : "index.docker.io/opensuse",
			"etcd": "quay.io/coreos/etcd"
		}
	}
	`

	testData := map[string][]string{
		"cool/stuff/busybox": []string{"https://index.docker.io", "flavio/busybox"},
		"cool/distro/leap":   []string{"https://index.docker.io", "opensuse/leap"},
		"etcd":               []string{"https://quay.io", "coreos/etcd"},
	}

	cfg := handleConfig(t, configData)
	runTranslateNameTestCase(t, cfg, testData)
}

func TestTranslateNameWithRootDefined(t *testing.T) {
	configData := `
	{
		"mappings" : {
			"cool/stuff":    "index.docker.io/flavio",
			"cool/distro":   "index.docker.io/opensuse",
			"etcd":          "quay.io/coreos/etcd",
			"/":             "registry.local.lan/docker.io"
		}
	}
	`

	testData := map[string][]string{
		"cool/stuff/busybox": []string{"https://registry.local.lan", "docker.io/cool/stuff/busybox"},
		"cool/distro/leap":   []string{"https://registry.local.lan", "docker.io/cool/distro/leap"},
		"etcd":               []string{"https://registry.local.lan", "docker.io/etcd"},
		"busybox":            []string{"https://registry.local.lan", "docker.io/busybox"},
	}

	mountsByReg := handleConfig(t, configData)
	runTranslateNameTestCase(t, mountsByReg, testData)
}

func handleConfig(t *testing.T, configData string) config.Rules {
	cfg, err := config.LoadConfig(configData)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	return cfg
}

func runTranslateNameTestCase(t *testing.T, rules config.Rules, testData map[string][]string) {
	for req := range testData {
		registry, remoteName, err := translateName(rules.Instance, req)
		if err != nil {
			t.Fatalf("Unexpected error while translating name of %s: %v", req, err)
		}
		if registry.String() != testData[req][0] {
			t.Fatalf(
				"Wrong registry mapping for %s: got '%s' instead of '%s'",
				req,
				registry,
				testData[req][0])
		}
		if remoteName != testData[req][1] {
			t.Fatalf(
				"Wrong remoteName mapping for %s: got '%s' instead of '%s'",
				req,
				remoteName,
				testData[req][1])
		}

	}
}
