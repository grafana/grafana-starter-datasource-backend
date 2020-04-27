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

	// Start listening to requests send from Grafana. This call is blocking so
	// it wont finish until Grafana shutsdown the process or the plugin choose
	// to exit close down by itself
	err := backend.Serve(backend.ServeOpts{
		QueryDataHandler:   ds,
		CheckHealthHandler: ds,
	})

	// Log any error if we could start the plugin.
	if err != nil {
		log.DefaultLogger.Error(err.Error())
		os.Exit(1)
	}
}
