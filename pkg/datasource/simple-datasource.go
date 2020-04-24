package datasource

import (
	"context"
	"math/rand"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

// SimpleDatasource is an example datasource used to scaffold
// new datasource plugins with an backend.
type SimpleDatasource struct{}

// New returns a new SimpleDatasource
func New() *SimpleDatasource {
	return &SimpleDatasource{}
}

// QueryData handles multiple queries and returns multiple responses.
// req contains the queries []DataQuery (where each query contains RefID as a unique identifer).
// The QueryDataResponse contains a map of RefID to the response for each query, and each response
// contains Frames ([]*Frame).
func (td *SimpleDatasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	log.DefaultLogger.Info("QueryData", "request", req)

	// create response struct
	response := backend.NewQueryDataResponse()

	// loop over queries and execute them individually.
	for _, q := range req.Queries {
		res, err := td.query(ctx, q)
		if err != nil {
			return nil, err
		}

		// save the response in a hashmap
		// based on with RefID as identifier
		response.Responses[q.RefID] = res
	}

	return response, nil
}

func (td *SimpleDatasource) query(ctx context.Context, query backend.DataQuery) (backend.DataResponse, error) {
	queryResponse := backend.DataResponse{}
	frame := data.NewFrame("an example result")
	frame.Fields = append(frame.Fields, data.NewField("", nil, []int64{1, 2}))
	queryResponse.Frames = append(queryResponse.Frames, frame)

	return queryResponse, nil
}

// CheckHealth handles health checks sent from Grafana to the plugin.
// The main use case for these health checks is the test button on the
// datasource configuration page which allows users to verify that
// a datasource is working as expected.
func (td *SimpleDatasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	var status = backend.HealthStatusOk
	var message = ""

	if rand.Int()%2 == 0 {
		status = backend.HealthStatusError
		message = "randomized error"
	}

	return &backend.CheckHealthResult{
		Status:  status,
		Message: message,
	}, nil
}
