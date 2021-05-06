package main

import (
	"os"

	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

func main() {
	instanceManager := datasource.NewAutoInstanceManager(NewSampleDatasource)
	err := datasource.Serve(datasource.ServeOpts{
		QueryDataHandler:    instanceManager,
		CheckHealthHandler:  instanceManager,
		CallResourceHandler: instanceManager,
		StreamHandler:       instanceManager,
	})
	// Log any error if we could start the plugin.
	if err != nil {
		log.DefaultLogger.Error(err.Error())
		os.Exit(1)
	}
}
