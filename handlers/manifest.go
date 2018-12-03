package handlers

import (
	"fmt"
	"net/http"

	"github.com/docker/distribution/registry/api/errcode"
	"github.com/docker/distribution/registry/api/v2"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func (app *App) GetManifest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	reference := vars["reference"]

	log.WithFields(log.Fields{
		"name":      name,
		"reference": reference,
		"host":      r.Host,
	}).Debug("GET manifest")

	rules := GetRulesByHost(r.Host, app.Rules)
	registry, remoteName, err := translateName(rules, name)
	if err != nil {
		errcode.ServeJSON(w, v2.ErrorCodeManifestUnknown)
		log.WithFields(log.Fields{
			"event": "translate name",
			"name":  name,
		}).Info(err)
		return
	}

	newUrl := fmt.Sprintf("%s/v2/%s/manifests/%s", registry.String(), remoteName, reference)
	log.WithFields(log.Fields{
		"event":       "redirect manifest",
		"name":        name,
		"reference":   reference,
		"redirectUrl": newUrl,
	}).Info("Redirecting manifest request")

	http.Redirect(w, r, newUrl, http.StatusTemporaryRedirect)
	return
}
