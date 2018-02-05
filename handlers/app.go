package handlers

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/flavio/collage/config"
	"github.com/flavio/collage/types"
)

type App struct {
	Mappings    map[string]string
	MountsByReg types.MountPointsByRegistry
}

func NewApp(cfg config.Config) (App, error) {
	app := App{}
	mountsByReg, err := getMountPointsByRegistry(cfg)

	if err != nil {
		return app, err
	}

	app.Mappings = cfg.Mappings
	app.MountsByReg = mountsByReg

	return app, nil
}

func getMountPointsByRegistry(c config.Config) (mappings types.MountPointsByRegistry, err error) {
	mappings = make(types.MountPointsByRegistry)

	for mount, source := range c.Mappings {
		mp := types.MountPoint{Target: mount}

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
			mappings[registry] = []types.MountPoint{}
		}
		mappings[registry] = append(mappings[registry], mp)
	}

	return
}
