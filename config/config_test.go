package config

import (
	"fmt"
	"net/url"
	"testing"
)

func TestParseSimpleConfig(t *testing.T) {
	configData := `
	{
		"mappings" : {
			"cool/stuff" : "index.docker.io/flavio",
			"cool/distro" : "index.docker.io/opensuse",
			"etcd": "quay.io/coreos/etcd"
		}
	}
	`

	rules := handleConfig(t, configData)
	fmt.Printf("%+v\n", rules.Instance.Mappings)
	fmt.Printf("%+v\n", rules.Instance.MountPointsByRegistry)

	for _, path := range []string{"cool/stuff", "cool/distro", "etcd"} {
		_, found := rules.Instance.Mappings[path]
		if !found {
			t.Fatalf("Not found %s", path)
		}
	}
}

func TestParseConfigWithRootMapping(t *testing.T) {
	configData := `
	{
		"mappings" : {
			"cool/stuff" : "index.docker.io/flavio",
			"cool/distro" : "index.docker.io/opensuse",
			"etcd": "quay.io/coreos/etcd",
			"/": "mirror.local.lan"
		},
		"vhosts" : {
			"docker-io-mirror.local.lan": {
				"mappings" : {
					"/" : "mirror.local.lan/docker.io",
					"cool/stuff" : "index.docker.io/flavio"
				}
			}
		}
	}
	`

	rules := handleConfig(t, configData)

	mappingsToCheck := []map[string]*url.URL{
		rules.Instance.Mappings,
		rules.Vhosts["docker-io-mirror.local.lan"].Mappings,
	}

	for _, mappings := range mappingsToCheck {
		_, found := mappings["/"]
		if !found {
			t.Fatal("Root mapping not found")
		}

		if len(mappings) != 1 {
			t.Fatalf("Wrong number of mappings found, expected 1 got %d", len(mappings))
		}
	}
}

func TestParseVHostConfig(t *testing.T) {
	configData := `
	{
		"vhosts" : {
			"docker-io-mirror.local.lan": {
				"mappings" : {
					"/" : "mirror.local.lan/docker.io"
				}
			},
			"quay-io-mirror.local.lan": {
				"mappings" : {
					"/" : "mirror.local.lan/quay.io"
				}
			}
		},
		"mappings" : {
			"cool/stuff" : "index.docker.io/flavio",
			"cool/distro" : "index.docker.io/opensuse",
			"etcd": "quay.io/coreos/etcd"
		}
	}
	`

	rules := handleConfig(t, configData)

	for _, path := range []string{"cool/stuff", "cool/distro", "etcd"} {
		_, found := rules.Instance.Mappings[path]
		if !found {
			t.Fatalf("Not found %s", path)
		}
	}

	for _, vhost := range []string{"docker-io-mirror.local.lan", "quay-io-mirror.local.lan"} {
		_, found := rules.Vhosts[vhost]
		if !found {
			t.Fatalf("vhost not found %s", vhost)
		}
	}

	dockerHubVHost := rules.Vhosts["docker-io-mirror.local.lan"]
	if _, found := dockerHubVHost.Mappings["/"]; !found {
		t.Fatal("no mapping found for docker hub vhost")
	}
}

func handleConfig(t *testing.T, configData string) Rules {
	rules, err := LoadConfig(configData)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	return rules
}
