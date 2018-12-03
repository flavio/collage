package handlers

import (
	"fmt"
	"net/http"

	"github.com/docker/distribution/registry/api/errcode"
	"github.com/docker/distribution/registry/api/v2"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func (app *App) GetLayer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	digest := vars["digest"]

	log.WithFields(log.Fields{
		"name":   name,
		"digest": digest,
		"host":   r.Host,
	}).Debug("GET layer")

	rules := GetRulesByHost(r.Host, app.Rules)
	registry, remoteName, err := translateName(rules, name)
	if err != nil {
		log.WithFields(log.Fields{
			"event": "translate name",
			"name":  name,
		}).Info(err)

		errcode.ServeJSON(w, v2.ErrorCodeBlobUnknown)
		return
	}

	newUrl := fmt.Sprintf("%s/v2/%s/blobs/%s", registry.String(), remoteName, digest)
	log.WithFields(log.Fields{
		"event":       "redirect blob",
		"name":        name,
		"digest":      digest,
		"redirectUrl": newUrl,
	}).Info("Redirecting pull layer request")

	http.Redirect(w, r, newUrl, http.StatusTemporaryRedirect)
	return
}
