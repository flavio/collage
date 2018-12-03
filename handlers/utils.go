package handlers

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/flavio/collage/config"
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
				target = mp.Target
				source = mp.Source
				break
			}
		}
		if target != "" {
			registry = reg
			break
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
