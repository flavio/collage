package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/flavio/collage/config"
	"github.com/flavio/collage/handlers"

	"github.com/docker/distribution/reference"
	gorilla_handlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/opencontainers/go-digest"
	log "github.com/sirupsen/logrus"
	"gopkg.in/urfave/cli.v1"
)

const VERSION = "0.1.0"

func main() {
	var port int
	var configFile, cert, key string
	var debug bool

	app := cli.NewApp()
	app.Name = "collage"
	app.Usage = "read-only registry made by repositories coming from multiple external registries"
	app.Version = VERSION

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "cert",
			Usage:       "Path to the certificate to use",
			EnvVar:      "COLLAGE_CERTIFICATE",
			Destination: &cert,
		},
		cli.StringFlag{
			Name:        "key",
			Usage:       "Path to the key to use",
			EnvVar:      "COLLAGE_KEY",
			Destination: &key,
		},
		cli.IntFlag{
			Name:        "port, p",
			Value:       5000,
			Usage:       "Listen to port",
			EnvVar:      "COLLAGE_PORT",
			Destination: &port,
		},
		cli.StringFlag{
			Name:        "config, c",
			Value:       "config.json",
			Usage:       "Configuration file",
			EnvVar:      "COLLAGE_CONFIG",
			Destination: &configFile,
		},
		cli.BoolFlag{
			Name:        "debug, d",
			Usage:       "Enable extra debugging",
			EnvVar:      "COLLAGE_DEBUG",
			Destination: &debug,
		},
	}

	app.Action = func(c *cli.Context) error {
		if debug {
			log.SetLevel(log.DebugLevel)
		}

		if (cert != "" && key == "") || (cert == "" && key != "") {
			return cli.NewExitError(
				errors.New("cert and key have to be specified at the same time"),
				1)
		}

		cfg, err := config.LoadConfig(configFile)
		if err != nil {
			return cli.NewExitError(err, 1)
		}

		hApp, err := handlers.NewApp(cfg)
		if err != nil {
			return cli.NewExitError(err, 1)
		}

		r := defineRoutes(hApp)
		loggedRouter := gorilla_handlers.LoggingHandler(os.Stdout, r)

		if cert == "" {
			fmt.Printf("Starting insecure server on :%d\n", port)
			err = http.ListenAndServe(fmt.Sprintf(":%d", port), loggedRouter)
		} else {
			fmt.Printf("Starting secure server on :%d\n", port)
			err = http.ListenAndServeTLS(
				fmt.Sprintf(":%d", port),
				cert,
				key,
				loggedRouter)
		}

		if err != nil {
			return cli.NewExitError(err, 1)
		}

		return nil
	}

	app.Run(os.Args)
}

func defineRoutes(app handlers.App) *mux.Router {
	r := mux.NewRouter()

	//GET CATALOG
	r.HandleFunc(
		"/v2/_catalog",
		app.GetCatalog).Methods("GET")

	//GET TAGS
	r.HandleFunc(
		"/v2/{name:"+reference.NameRegexp.String()+"}/tags/list",
		app.GetRepositoryTags).Methods("GET", "HEAD")

	//GET MANIFEST
	r.HandleFunc(
		"/v2/{name:"+reference.NameRegexp.String()+"}/manifests/{reference:"+reference.TagRegexp.String()+"|"+digest.DigestRegexp.String()+"}",
		app.GetManifest).Methods("GET", "HEAD")

	//GET LAYER
	r.HandleFunc(
		"/v2/{name:"+reference.NameRegexp.String()+"}/blobs/{digest:"+digest.DigestRegexp.String()+"}",
		app.GetLayer).Methods("GET", "HEAD")

	return r
}
