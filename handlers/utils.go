package handlers

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	docker_types "github.com/docker/docker/api/types"
	"github.com/flavio/collage/config"
	"github.com/genuinetools/reg/registry"
)

func translateName(mappingRules config.MappingRules, name string) (registry *url.URL, remoteName string, err error) {
	var source, target string

	if _, rootMappingDefined := mappingRules.Mappings["/"]; rootMappingDefined {
		// '/' has the precedence over everything, the config is also
		// purged of all the other mappings -> we have just one registry with
		// one mount point
		for reg, mountPoints := range mappingRules.MountPointsByRegistry {
			registry = reg
			for _, mp := range mountPoints {
				remoteName = fmt.Sprintf("%s/%s", mp.Source, name)
				return
			}
		}
	}

	for reg, mountPoints := range mappingRules.MountPointsByRegistry {
		for _, mp := range mountPoints {
			if strings.HasPrefix(name, mp.Target) {
				// The match with the longest length has precedence because
				// is more specific
				if len(mp.Target) > len(target) {
					target = mp.Target
					source = mp.Source
					registry = reg
				}
			}
		}
	}
	if registry == nil || target == "" || source == "" {
		err = errors.New("Cannot find mount point")
		return
	}

	remoteName = strings.Replace(name, target, source, 1)

	return
}

func GetRulesByHost(host string, rules config.Rules) config.MappingRules {
	if _, found := rules.Vhosts[host]; found {
		return rules.Vhosts[host]
	}
	return rules.Instance
}

func NewRegistry(url *url.URL, cfg *config.Config) (*registry.Registry, error) {
	auth := docker_types.AuthConfig{
		ServerAddress: url.String(),
	}

	opt := registry.Opt{
		Debug: cfg.Debug,
	}

	if url.Scheme == "https" && cfg.CertPool != nil {
		// We have to inject our custom certificates
		transport := &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
				RootCAs:            cfg.CertPool,
			},
		}
		opt.Insecure = true

		reg, err := registry.New(auth, opt)
		if err != nil {
			return nil, err
		}
		reg.Client.Transport = transport

		return reg, nil
	}

	return registry.New(auth, opt)
}
