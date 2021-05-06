package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/grafana/grafana-plugin-sdk-go/live"
)

// newDatasource returns datasource.ServeOpts.
func newDatasource() datasource.ServeOpts {
	// creates a instance manager for your plugin. The function passed
	// into `NewInstanceManger` is called when the instance is created
	// for the first time or when a datasource configuration changed.
	im := datasource.NewInstanceManager(newDataSourceInstance)
	ds := &SampleDatasource{
		im: im,
	}

	return datasource.ServeOpts{
		QueryDataHandler:   ds,
		CheckHealthHandler: ds,
		StreamHandler:      ds,
	}
}

// SampleDatasource is an example datasource used to scaffold
// new datasource plugins with an backend.
type SampleDatasource struct {
	// The instance manager can help with lifecycle management
	// of datasource instances in plugins. It's not a requirements
	// but a best practice that we recommend that you follow.
	im instancemgmt.InstanceManager
}

func (td *SampleDatasource) getInstance(pCtx backend.PluginContext) (*datasourceInstance, error) {
	instance, err := td.im.Get(pCtx)
	if err != nil {
		return nil, err
	}
	dsInstance, ok := instance.(*datasourceInstance)
	if !ok {
		return nil, fmt.Errorf("malformed instance type %T", instance)
	}
	return dsInstance, nil
}

// QueryData handles multiple queries and returns multiple responses.
// req contains the queries []DataQuery (where each query contains RefID as a unique identifier).
// The QueryDataResponse contains a map of RefID to the response for each query, and each response
// contains Frames ([]*Frame).
func (td *SampleDatasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	log.DefaultLogger.Info("QueryData called", "request", req)

	_, err := td.getInstance(req.PluginContext)
	if err != nil {
		return nil, err
	}

	// create response struct
	response := backend.NewQueryDataResponse()

	// loop over queries and execute them individually.
	for _, q := range req.Queries {
		res := td.query(ctx, req.PluginContext, q)

		// save the response in a hashmap
		// based on with RefID as identifier
		response.Responses[q.RefID] = res
	}

	return response, nil
}

type queryModel struct {
	WithStreaming bool `json:"withStreaming"`
}

func (td *SampleDatasource) query(ctx context.Context, pCtx backend.PluginContext, query backend.DataQuery) backend.DataResponse {
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
func (td *SampleDatasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
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

func (td *SampleDatasource) SubscribeStream(ctx context.Context, req *backend.SubscribeStreamRequest) (*backend.SubscribeStreamResponse, error) {
	log.DefaultLogger.Info("SubscribeStream called", "request", req)

	return &backend.SubscribeStreamResponse{
		Status: backend.SubscribeStreamStatusOK,
		// Enabling UseRunStream will make Grafana open a unidirectional stream
		// to consume data from a plugin while active subscribers exist.
		UseRunStream: true,
	}, nil
}

func (td *SampleDatasource) PublishStream(ctx context.Context, req *backend.PublishStreamRequest) (*backend.PublishStreamResponse, error) {
	log.DefaultLogger.Info("PublishStream called", "request", req)

	return &backend.PublishStreamResponse{
		Status: backend.PublishStreamStatusPermissionDenied,
	}, nil
}

func (td *SampleDatasource) RunStream(ctx context.Context, req *backend.RunStreamRequest, sender backend.StreamPacketSender) error {
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

type datasourceSettings struct{}

type datasourceInstance struct {
	settings datasourceSettings
}

func newDataSourceInstance(setting backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	log.DefaultLogger.Info("Creating new instance", "updated", setting.Updated)
	var dsSettings datasourceSettings
	if err := json.Unmarshal(setting.JSONData, &dsSettings); err != nil {
		return nil, err
	}
	return &datasourceInstance{
		settings: dsSettings,
	}, nil
}

func (s *datasourceInstance) Dispose() {
	// Called before creating a new instance to allow plugin authors
	// to cleanup.
}
