package main

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

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

func cleanup(socket string) {
	if socket != "" {
		os.Remove(socket)
	}
}

func main() {
	var port int
	var configData, configFile, cert, key, socket string
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
			Name:        "socket",
			Usage:       "Bind to a unix socket",
			EnvVar:      "COLLAGE_SOCKET",
			Destination: &socket,
		},
		cli.StringFlag{
			Name:        "config-file",
			Value:       "",
			Usage:       "Configuration file",
			EnvVar:      "COLLAGE_CONFIG_FILE",
			Destination: &configFile,
		},
		cli.StringFlag{
			Name:        "config, c",
			Value:       "",
			Usage:       "JSON configuration",
			EnvVar:      "COLLAGE_CONFIG",
			Destination: &configData,
		},
		cli.BoolFlag{
			Name:        "debug, d",
			Usage:       "Enable extra debugging",
			EnvVar:      "COLLAGE_DEBUG",
			Destination: &debug,
		},
	}

	app.Action = func(c *cli.Context) error {

		// ensure cleanup is done when the program is terminated
		sigChannel := make(chan os.Signal)
		signal.Notify(sigChannel, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-sigChannel
			cleanup(socket)
			os.Exit(0)
		}()

		if debug {
			log.SetLevel(log.DebugLevel)
		}

		if (cert != "" && key == "") || (cert == "" && key != "") {
			return cli.NewExitError(
				errors.New("cert and key have to be specified at the same time"),
				1)
		}

		if configFile != "" && configData != "" {
			return cli.NewExitError(
				errors.New("'--config-file' and '--config' cannot be specified at the same time"),
				1)
		}

		if configFile == "" && configData == "" {
			return cli.NewExitError(
				errors.New("You must specify either '--config-file' or '--config'"),
				1)
		}

		var cfg config.Config
		var err error

		if configFile != "" {
			cfg, err = config.LoadConfigFromFile(configFile)
		} else {
			cfg, err = config.LoadConfig(configData)
		}
		if err != nil {
			return cli.NewExitError(err, 1)
		}

		hApp, err := handlers.NewApp(cfg)
		if err != nil {
			return cli.NewExitError(err, 1)
		}

		r := defineRoutes(hApp)
		loggedRouter := gorilla_handlers.LoggingHandler(os.Stdout, r)

		if socket != "" {
			_, err := os.Stat(socket)
			if err == nil {
				return cli.NewExitError(
					fmt.Errorf(
						"Cannot create socket %q, file already in place. Is another instance already running?",
						socket),
					1)
			} else if !os.IsNotExist(err) {
				return cli.NewExitError(err, 1)
			}
			unixListener, err := net.Listen("unix", socket)
			if err != nil {
				return cli.NewExitError(err, 1)
			}
			defer func() {
				unixListener.Close()
				os.Remove(socket)
			}()
			err = http.Serve(unixListener, loggedRouter)
		} else {
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
