package handlers

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/docker/distribution/registry/api/errcode"
	"github.com/flavio/collage/config"
	log "github.com/sirupsen/logrus"
)

func (app *App) GetApiVersionCheck(w http.ResponseWriter, r *http.Request) {
	log.WithFields(log.Fields{
		"host": r.Host,
	}).Debug("GET /v2/")

	if _, hasVhost := app.Rules.Vhosts[r.Host]; hasVhost {
		if _, hasRootMapping := app.Rules.Vhosts[r.Host].Mappings["/"]; hasRootMapping {
			// this vhost has a root mapping in place -> there's going to be only
			// a single registry for it
			var realRegistry *url.URL
			for reg := range app.Rules.Vhosts[r.Host].MountPointsByRegistry {
				realRegistry = reg
			}

			bearerRealm, err := getRegistryBearerRealm(realRegistry, app.Rules)
			if err != nil {
				log.WithFields(log.Fields{
					"event":        "getRegistryBearerRealm",
					"realRegistry": realRegistry.String(),
				}).Error(err)
				errcode.ServeJSON(w, errcode.ErrorCodeUnknown.WithDetail(err))
				return
			}

			if bearerRealm != "" {
				service := fmt.Sprintf("collage;%s", r.Host)

				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.Header().Set("Docker-Distribution-Api-Version", "registry/2.0")
				w.Header().Set(
					"WWW-Authenticate",
					fmt.Sprintf("Bearer realm=\"%s\",service=\"%s\"", bearerRealm, service))
				w.WriteHeader(http.StatusUnauthorized)
				io.WriteString(w, `{"errors":[{"code":"UNAUTHORIZED","message":"authentication required","detail":null}]}`)
			} else {
				newUrl := fmt.Sprintf("http://%s/v2/", realRegistry)
				http.Redirect(w, r, newUrl, http.StatusTemporaryRedirect)
			}

			return
		}
	}

	// we can't redirect to another registry
	// let's answer we support v2 protocol
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Docker-Distribution-Api-Version", "registry/2.0")
	io.WriteString(w, "{}")
}

func getRegistryBearerRealm(registry *url.URL, rules config.Rules) (string, error) {
	bearer, known := rules.RegistryBearerRealm[registry]
	if known {
		return bearer, nil
	}

	apiEndpoint := fmt.Sprintf("%s/v2/", registry.String())
	resp, err := http.Get(apiEndpoint)
	if err != nil {
		return "", err
	}

	if resp.StatusCode == http.StatusUnauthorized {
		auth := resp.Header[http.CanonicalHeaderKey("WWW-Authenticate")]
		if len(auth) == 1 {
			for _, info := range strings.Split(auth[0], ",") {
				if strings.HasPrefix(info, "Bearer realm=") {
					rules.RegistryBearerRealm[registry] = strings.TrimSuffix(strings.TrimPrefix(info, "Bearer realm=\""), "\"")
					return rules.RegistryBearerRealm[registry], nil
				}
			}
		}
		return "", fmt.Errorf(
			"%s/v2/ didn't provide authentication informations %+v",
			apiEndpoint,
			resp.Header)
	}

	rules.RegistryBearerRealm[registry] = ""
	return rules.RegistryBearerRealm[registry], nil
}
