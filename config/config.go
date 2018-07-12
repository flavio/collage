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

type VHostSettings struct {
	InstanceMappings map[string]string `json:"mappings"`
}

type Config struct {
	InstanceMappings     map[string]string        `json:"mappings"`
	VirtualHostsMappings map[string]VHostSettings `json:"vhosts"`
}

type MappingRules struct {
	Mappings              map[string]string
	MountPointsByRegistry map[string][]MountPoint
}

type Rules struct {
	Instance MappingRules
	Vhosts   map[string]MappingRules
}

func validateMappings(mappings map[string]string) {
	if _, rootDefined := mappings["/"]; !rootDefined {
		return
	}
	log.Warn(
		"A mapping for root ('/') is defined, ignoring all the",
		" other directives because it has the highest priority")
	routesToDelete := []string{}
	for route := range mappings {
		if route != "/" {
			routesToDelete = append(routesToDelete, route)
		}
	}

	for _, route := range routesToDelete {
		delete(mappings, route)
	}
}

func (c *Config) Validate() {
	validateMappings(c.InstanceMappings)
	for vhost := range c.VirtualHostsMappings {
		validateMappings(c.VirtualHostsMappings[vhost].InstanceMappings)
	}
}

func getMountPointsByRegistry(mappings map[string]string) (mappingsByRegistry map[string][]MountPoint, err error) {
	mappingsByRegistry = make(map[string][]MountPoint)

	for mount, source := range mappings {
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

		if _, hasKey := mappingsByRegistry[registry]; !hasKey {
			mappingsByRegistry[registry] = []MountPoint{}
		}
		mappingsByRegistry[registry] = append(mappingsByRegistry[registry], mp)
	}

	return
}

func (cfg *Config) Rules() (rules Rules, err error) {
	cfg.Validate()

	rules.Instance.Mappings = cfg.InstanceMappings

	rules.Instance.MountPointsByRegistry, err = getMountPointsByRegistry(cfg.InstanceMappings)
	if err != nil {
		return Rules{}, err
	}
	rules.Vhosts = make(map[string]MappingRules)

	for vhost := range cfg.VirtualHostsMappings {
		var mappingRules MappingRules
		mappingRules.Mappings = cfg.VirtualHostsMappings[vhost].InstanceMappings

		mappingRules.MountPointsByRegistry, err = getMountPointsByRegistry(mappingRules.Mappings)
		if err != nil {
			return Rules{}, err
		}

		rules.Vhosts[vhost] = mappingRules
	}

	return
}

func LoadConfig(data string) (rules Rules, err error) {
	var cfg Config

	decoder := json.NewDecoder(strings.NewReader(data))
	err = decoder.Decode(&cfg)
	if err != nil {
		return
	}

	return cfg.Rules()
}

func LoadConfigFromFile(path string) (rules Rules, err error) {
	var cfg Config

	file, err := os.Open(path)
	if err != nil {
		return
	}

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&cfg)
	if err != nil {
		return
	}

	return cfg.Rules()
}
