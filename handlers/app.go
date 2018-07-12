package handlers

import (
	"github.com/flavio/collage/config"
	log "github.com/sirupsen/logrus"
)

type App struct {
	Config config.Config
}

func NewApp(cfg config.Config) (App, error) {
	app := App{Config: cfg}

	log.WithFields(log.Fields{
		"Mappings":           app.Config.Mappings,
		"MappingsByRegistry": app.Config.MappingsByRegistry,
	}).Info("app initiated")

	return app, nil
}
