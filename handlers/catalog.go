package handlers

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/flavio/collage/config"
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
		upstreamCatalog, err := upstreamCatalog(registry, app.Cfg)
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

func upstreamCatalog(registryUrl *url.URL, cfg *config.Config) ([]string, error) {
	reg, err := NewRegistry(registryUrl, cfg)
	if err != nil {
		return []string{}, err
	}

	return reg.Catalog("")
}
