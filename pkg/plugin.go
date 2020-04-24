package main

import (
	"os"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/test-datasource/pkg/datasource"
)

func main() {
	// creates a plugin instance
	ds := datasource.New()

	// start serving plugin requests from grafana-server
	err := backend.Serve(backend.ServeOpts{
		QueryDataHandler:   ds,
		CheckHealthHandler: ds,
	})

	if err != nil {
		log.DefaultLogger.Error(err.Error())
		os.Exit(1)
	}
}
