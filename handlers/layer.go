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
	}).Debug("GET layer")

	registry, remoteName, err := translateName(app.Config, name)
	if err != nil {
		log.WithFields(log.Fields{
			"event": "translate name",
			"name":  name,
		}).Info(err)

		errcode.ServeJSON(w, v2.ErrorCodeBlobUnknown)
		return
	}

	newUrl := fmt.Sprintf("https://%s/v2/%s/blobs/%s", registry, remoteName, digest)
	log.WithFields(log.Fields{
		"event":       "redirect blob",
		"name":        name,
		"digest":      digest,
		"redirectUrl": newUrl,
	}).Info("Redirecting pull layer request")

	http.Redirect(w, r, newUrl, http.StatusTemporaryRedirect)
	return
}
