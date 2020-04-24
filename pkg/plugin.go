package main

import (
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/test-datasource/pkg/datasource"
)

func main() {
	backend.SetupPluginEnvironment("test-datasource")

	pluginLogger := log.New()
	ds := datasource.New()

	err := backend.Serve(backend.ServeOpts{
		QueryDataHandler:   ds,
		CheckHealthHandler: ds,
	})

	if err != nil {
		pluginLogger.Error(err.Error())
	}
}
