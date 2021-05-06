package main

import (
	"os"

	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

func main() {
	err := datasource.Manage(NewSampleDatasource, datasource.ManageOpts{})
	// Log any error if we could start the plugin.
	if err != nil {
		log.DefaultLogger.Error(err.Error())
		os.Exit(1)
	}
}
