package handlers

import (
	"errors"
	"fmt"
	"strings"

	"github.com/flavio/collage/config"
)

func translateName(cfg config.Config, name string) (registry string, remoteName string, err error) {
	var source, target string

	if _, rootMappingDefined := cfg.Mappings["/"]; rootMappingDefined {
		// '/' has the precedence over everything, the config is also
		// purged of all the other mappings -> we have just one registry with
		// one mount point
		for reg, mountPoints := range cfg.MappingsByRegistry {
			registry = reg
			for _, mp := range mountPoints {
				remoteName = fmt.Sprintf("%s/%s", mp.Source, name)
				return
			}
		}
	}

	for reg, mountPoints := range cfg.MappingsByRegistry {
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
	if registry == "" || target == "" || source == "" {
		err = errors.New("Cannot find mount point")
		return
	}

	remoteName = strings.Replace(name, target, source, 1)

	return
}
