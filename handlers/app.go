package handlers

import (
	"github.com/flavio/collage/config"
	log "github.com/sirupsen/logrus"
)

type App struct {
	Cfg   *config.Config
	Rules config.Rules
}

func NewApp(cfg *config.Config) (App, error) {
	rules, err := cfg.Rules()
	if err != nil {
		return App{}, err
	}

	app := App{
		Rules: rules,
		Cfg:   cfg,
	}

	log.WithFields(log.Fields{
		"Rules":  app.Rules,
		"Config": app.Cfg,
	}).Info("app initiated")

	return app, nil
}
