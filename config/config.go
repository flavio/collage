package config

import (
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

type VHostSettings struct {
	InstanceMappings map[string]string `json:"mappings"`
}

type Config struct {
	InstanceMappings     map[string]string        `json:"mappings"`
	VirtualHostsMappings map[string]VHostSettings `json:"vhosts"`

	// This are attributes that are not part of the JSON configuration
	Debug    bool
	CertPool *x509.CertPool
}

// Defines a mount point rule.
// Given `cool/stuff` -> `index.docker.io/flavio`:
//   * Source is going to be index.docker.io/flavio
//   * Target is going to be `cool/stuff`
type MountPoint struct {
	Source string
	Target string
}

type MappingRules struct {
	Mappings              map[string]*url.URL
	MountPointsByRegistry map[*url.URL][]MountPoint
}

type Rules struct {
	Instance            MappingRules
	Vhosts              map[string]MappingRules
	RegistryBearerRealm map[*url.URL]string
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

func getMountPointsByRegistry(mappings map[string]*url.URL) (mappingsByRegistry map[*url.URL][]MountPoint, err error) {
	mappingsByRegistry = make(map[*url.URL][]MountPoint)

	for mount, source := range mappings {
		mp := MountPoint{Target: mount}

		registry, _ := url.Parse(fmt.Sprintf("%s://%s", source.Scheme, source.Host))
		if strings.HasPrefix(source.Path, "/") {
			mp.Source = source.Path[1:]
		} else {
			mp.Source = source.Path
		}

		if _, hasKey := mappingsByRegistry[registry]; !hasKey {
			mappingsByRegistry[registry] = []MountPoint{}
		}
		mappingsByRegistry[registry] = append(mappingsByRegistry[registry], mp)
	}

	return
}

func parseMappings(mappings map[string]string) (ret map[string]*url.URL, err error) {
	ret = make(map[string]*url.URL)

	for target, source := range mappings {
		if !strings.HasPrefix(source, "https://") && !strings.HasPrefix(source, "http://") {
			source = fmt.Sprintf("https://%s", source)
		}

		url, err2 := url.Parse(source)
		if err2 != nil {
			err = fmt.Errorf("Cannot parse url %s: %v", source, err2)
			return
		}

		ret[target] = url
	}

	return
}

func (cfg *Config) Rules() (rules Rules, err error) {
	cfg.Validate()

	rules.Instance.Mappings, err = parseMappings(cfg.InstanceMappings)
	if err != nil {
		return Rules{}, err
	}

	rules.Instance.MountPointsByRegistry, err = getMountPointsByRegistry(rules.Instance.Mappings)
	if err != nil {
		return Rules{}, err
	}
	rules.Vhosts = make(map[string]MappingRules)

	for vhost := range cfg.VirtualHostsMappings {
		var mappingRules MappingRules
		mappingRules.Mappings, err = parseMappings(cfg.VirtualHostsMappings[vhost].InstanceMappings)
		if err != nil {
			return Rules{}, err
		}

		mappingRules.MountPointsByRegistry, err = getMountPointsByRegistry(mappingRules.Mappings)
		if err != nil {
			return Rules{}, err
		}

		rules.Vhosts[vhost] = mappingRules
	}

	rules.RegistryBearerRealm = make(map[*url.URL]string)

	return
}

func NewFromData(data string) (cfg Config, err error) {
	decoder := json.NewDecoder(strings.NewReader(data))
	err = decoder.Decode(&cfg)
	if err != nil {
		return
	}

	return
}

func NewFromFile(path string) (cfg Config, err error) {
	file, err := os.Open(path)
	if err != nil {
		return
	}

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&cfg)
	if err != nil {
		return
	}

	return
}

func (cfg *Config) LoadExtraCerts(dir string) error {
	if dir == "" {
		return nil
	}

	// get system pool certificates
	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}

	matches, err := filepath.Glob(path.Join(dir, "*.pem"))
	if err != nil {
		return err
	}
	if len(matches) == 0 {
		log.WithFields(
			log.Fields{
				"dir": dir,
			}).Debug("No certificates found")
	}

	for _, match := range matches {
		cert, err := ioutil.ReadFile(match)
		if err != nil {
			return fmt.Errorf(
				"Error while loading certificate %s: %v",
				match,
				err)
		}

		// Append cert to the system pool
		if ok := rootCAs.AppendCertsFromPEM(cert); ok {
			log.WithFields(log.Fields{
				"cert": match,
			}).Debug("Extra certificate imported")
		}
	}
	cfg.CertPool = rootCAs

	return nil
}
