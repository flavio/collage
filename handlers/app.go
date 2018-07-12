package handlers

import (
	"github.com/flavio/collage/config"
	log "github.com/sirupsen/logrus"
)

type App struct {
	Rules config.Rules
}

func NewApp(rules config.Rules) (App, error) {
	app := App{Rules: rules}

	log.WithFields(log.Fields{
		"Rules": app.Rules,
	}).Info("app initiated")

	return app, nil
}
