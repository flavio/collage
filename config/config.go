package config

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

type MountPoint struct {
	Source string
	Target string
}

type MountPointsByRegistry map[string][]MountPoint

type Config struct {
	Mappings           map[string]string `json:mappings`
	MappingsByRegistry MountPointsByRegistry
}

func (c *Config) Validate() {
	if _, rootDefined := c.Mappings["/"]; rootDefined {
		log.Warn(
			"A mapping for root ('/') is defined, ignoring all the",
			" other directives because it has the highest priority")
		routesToDelete := []string{}
		for route := range c.Mappings {
			if route != "/" {
				routesToDelete = append(routesToDelete, route)
			}
		}

		for _, route := range routesToDelete {
			delete(c.Mappings, route)
		}
	}
}

func getMountPointsByRegistry(c Config) (mappings MountPointsByRegistry, err error) {
	mappings = make(MountPointsByRegistry)

	for mount, source := range c.Mappings {
		mp := MountPoint{Target: mount}

		if !strings.HasPrefix(source, "https://") && !strings.HasPrefix(source, "http://") {
			source = fmt.Sprintf("https://%s", source)
		}

		u, err2 := url.Parse(source)
		if err2 != nil {
			err = fmt.Errorf("Cannot parse url %s: %v", source, err2)
			return
		}

		registry := u.Host
		if strings.HasPrefix(u.Path, "/") {
			mp.Source = u.Path[1:]
		} else {
			mp.Source = u.Path
		}

		if _, hasKey := mappings[registry]; !hasKey {
			mappings[registry] = []MountPoint{}
		}
		mappings[registry] = append(mappings[registry], mp)
	}

	return
}
func LoadConfig(data string) (cfg Config, err error) {
	decoder := json.NewDecoder(strings.NewReader(data))
	err = decoder.Decode(&cfg)
	if err != nil {
		return
	}

	cfg.Validate()
	cfg.MappingsByRegistry, err = getMountPointsByRegistry(cfg)
	return
}

func LoadConfigFromFile(path string) (cfg Config, err error) {
	file, err := os.Open(path)
	if err != nil {
		return
	}

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&cfg)
	if err != nil {
		return
	}

	cfg.Validate()
	cfg.MappingsByRegistry, err = getMountPointsByRegistry(cfg)
	return
}
