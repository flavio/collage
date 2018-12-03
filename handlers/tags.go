package handlers

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/docker/distribution/registry/api/errcode"
	"github.com/docker/distribution/registry/api/v2"
	docker_types "github.com/docker/docker/api/types"
	"github.com/genuinetools/reg/registry"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type tagsResponse struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

func (app *App) GetRepositoryTags(w http.ResponseWriter, r *http.Request) {
	//TODO: handle pagination?

	vars := mux.Vars(r)
	name := vars["name"]

	log.WithFields(log.Fields{
		"name": name,
		"host": r.Host,
	}).Debug("GetRepositoryTags")

	rules := GetRulesByHost(r.Host, app.Rules)
	registry, remoteName, err := translateName(rules, name)
	if err != nil {
		log.WithFields(log.Fields{
			"event": "translate name",
			"name":  name,
		}).Info(err)
		errcode.ServeJSON(w, v2.ErrorCodeNameUnknown)
		return
	}

	tags, err := upstreamTags(registry, remoteName)
	if err != nil {
		log.WithFields(log.Fields{
			"event":    "upstreamTags",
			"registry": registry,
			"name":     remoteName,
		}).Error(err)
		errcode.ServeJSON(w, errcode.ErrorCodeUnknown.WithDetail(err))
		return
	}

	log.WithFields(log.Fields{
		"registry":   registry,
		"repository": remoteName,
		"tags":       tags,
	}).Debug("Remote tags")

	response := tagsResponse{
		Name: name,
		Tags: tags,
	}

	json.NewEncoder(w).Encode(response)
}

func upstreamTags(registryUrl *url.URL, name string) (tags []string, err error) {
	auth := docker_types.AuthConfig{
		ServerAddress: registryUrl.String(),
	}

	reg, err := registry.New(auth, registry.Opt{})
	if err != nil {
		return
	}

	return reg.Tags(name)
}
