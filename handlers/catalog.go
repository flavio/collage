package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	docker_types "github.com/docker/docker/api/types"
	"github.com/jessfraz/reg/registry"
	log "github.com/sirupsen/logrus"
)

type catalogResponse struct {
	Repositories []string `json:"repositories"`
}

func (app *App) GetCatalog(w http.ResponseWriter, r *http.Request) {
	log.WithFields(log.Fields{
		"event": "fetch catalog",
		"host":  r.Host,
	}).Debug("GetCatalog")

	//TODO: handle pagination?
	response := catalogResponse{}

	rules := GetRulesByHost(r.Host, app.Rules)

	for registry, mounts := range rules.MountPointsByRegistry {
		upstreamCatalog, err := upstreamCatalog(registry)
		if err != nil {
			log.WithFields(log.Fields{
				"event":    "fetch catalog",
				"registry": registry,
				"host":     r.Host,
			}).Error(err)
		}

		for _, remoteRepo := range upstreamCatalog {
			for _, mount := range mounts {
				if strings.HasPrefix(remoteRepo, mount.Source) {
					response.Repositories = append(
						response.Repositories,
						strings.Replace(remoteRepo, mount.Source, mount.Target, 1))
				}
			}
		}
	}

	json.NewEncoder(w).Encode(response)
}

func upstreamCatalog(registryUrl string) (repositories []string, err error) {
	auth := docker_types.AuthConfig{
		ServerAddress: registryUrl,
	}

	reg, err := registry.New(auth, false)
	if err != nil {
		return
	}

	return reg.Catalog("")
}
