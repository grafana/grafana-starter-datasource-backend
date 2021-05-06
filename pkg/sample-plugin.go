package main

import (
	"context"
	"encoding/json"
	"errors"
	"math/rand"
	"strconv"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/grafana/grafana-plugin-sdk-go/live"
)

// NewSampleDatasource creates new datasource instance.
func NewSampleDatasource(_ backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	return &SampleDatasource{
		closeCh: make(chan struct{}),
	}, nil
}

// SampleDatasource is an example datasource used to scaffold
// new datasource plugins with an backend.
type SampleDatasource struct {
	closeCh chan struct{}
}

// Dispose here tells plugin SDK that plugin wants to clean up resources
// when new instance created. As soon as datasource settings change detected
// by SDK old datasource instance will be disposed and new one will be created
// using NewSampleDatasource.
func (d *SampleDatasource) Dispose() {
	close(d.closeCh)
}

// QueryData handles multiple queries and returns multiple responses.
// req contains the queries []DataQuery (where each query contains RefID as a unique identifier).
// The QueryDataResponse contains a map of RefID to the response for each query, and each response
// contains Frames ([]*Frame).
func (d *SampleDatasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	log.DefaultLogger.Info("QueryData called", "request", req)

	// create response struct
	response := backend.NewQueryDataResponse()

	// loop over queries and execute them individually.
	for _, q := range req.Queries {
		res := d.query(ctx, req.PluginContext, q)

		// save the response in a hashmap
		// based on with RefID as identifier
		response.Responses[q.RefID] = res
	}

	return response, nil
}

type queryModel struct {
	WithStreaming bool `json:"withStreaming"`
}

func (d *SampleDatasource) query(_ context.Context, pCtx backend.PluginContext, query backend.DataQuery) backend.DataResponse {
	// Unmarshal the json into our queryModel
	var qm queryModel

	response := backend.DataResponse{}

	response.Error = json.Unmarshal(query.JSON, &qm)
	if response.Error != nil {
		return response
	}

	// create data frame response
	frame := data.NewFrame("response")

	// add the time dimension
	frame.Fields = append(frame.Fields,
		data.NewField("time", nil, []time.Time{query.TimeRange.From, query.TimeRange.To}),
	)

	// add values
	frame.Fields = append(frame.Fields,
		data.NewField("values", nil, []int64{10, 20}),
	)

	// If datasource created with streaming on then return a channel
	// to subscribe on client-side and consume updated from a plugin.
	if qm.WithStreaming {
		channel := live.Channel{
			Scope:     live.ScopeDatasource,
			Namespace: strconv.FormatInt(pCtx.DataSourceInstanceSettings.ID, 10),
			Path:      "stream",
		}
		frame.SetMeta(&data.FrameMeta{Channel: channel.String()})
	}

	// add the frames to the response
	response.Frames = append(response.Frames, frame)

	return response
}

// CheckHealth handles health checks sent from Grafana to the plugin.
// The main use case for these health checks is the test button on the
// datasource configuration page which allows users to verify that
// a datasource is working as expected.
func (d *SampleDatasource) CheckHealth(_ context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	log.DefaultLogger.Info("CheckHealth called", "request", req)

	var status = backend.HealthStatusOk
	var message = "Data source is working"

	if rand.Int()%2 == 0 {
		status = backend.HealthStatusError
		message = "randomized error"
	}

	return &backend.CheckHealthResult{
		Status:  status,
		Message: message,
	}, nil
}

func (d *SampleDatasource) SubscribeStream(_ context.Context, req *backend.SubscribeStreamRequest) (*backend.SubscribeStreamResponse, error) {
	log.DefaultLogger.Info("SubscribeStream called", "request", req)

	return &backend.SubscribeStreamResponse{
		Status: backend.SubscribeStreamStatusOK,
		// Enabling UseRunStream will make Grafana open a unidirectional stream
		// to consume data from a plugin while active subscribers exist.
		UseRunStream: true,
	}, nil
}

func (d *SampleDatasource) PublishStream(_ context.Context, req *backend.PublishStreamRequest) (*backend.PublishStreamResponse, error) {
	log.DefaultLogger.Info("PublishStream called", "request", req)

	return &backend.PublishStreamResponse{
		Status: backend.PublishStreamStatusPermissionDenied,
	}, nil
}

func (d *SampleDatasource) RunStream(ctx context.Context, req *backend.RunStreamRequest, sender backend.StreamPacketSender) error {
	log.DefaultLogger.Info("RunStream called", "request", req)

	// Create the same data frame as for query data.
	frame := data.NewFrame("response")

	// Add the time dimension.
	frame.Fields = append(frame.Fields,
		data.NewField("time", nil, make([]time.Time, 1)),
	)
	// Add values dimension.
	frame.Fields = append(frame.Fields,
		data.NewField("values", nil, make([]int64, 1)),
	)

	counter := 0

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-d.closeCh:
			log.DefaultLogger.Info("Datasource restart")
			return errors.New("datasource closed")
		case <-time.After(200 * time.Millisecond):
			// Send new data periodically.
			frame.Fields[0].Set(0, time.Now())
			frame.Fields[1].Set(0, int64(10*(counter%2+1)))

			counter++

			frameJSON, err := json.Marshal(frame)
			if err != nil {
				log.DefaultLogger.Error("Error marshaling frame", "error", err)
				continue
			}

			err = sender.Send(&backend.StreamPacket{
				Data: frameJSON,
			})
			if err != nil {
				log.DefaultLogger.Error("Error sending frame", "error", err)
				continue
			}
		}
	}
}
