package datasource

import (
	"context"

	"github.com/grafana/grafana-plugin-sdk-go/data"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
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
	qdr := backend.NewQueryDataResponse()

	for _, q := range req.Queries {
		res := backend.DataResponse{}
		frame := data.NewFrame("an example result")
		frame.Fields = append(frame.Fields, data.NewField("", nil, []int64{1, 2}))
		res.Frames = append(res.Frames, frame)

		qdr.Responses[q.RefID] = res
	}

	return qdr, nil
}

// CheckHealth handles health checks sent from Grafana to the plugin.
// The main use case for these health checks is the test button on the
// datasource configuration page which allows users to verify that
// a datasource is working as expected.
func (td *SimpleDatasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	return &backend.CheckHealthResult{}, nil
}
