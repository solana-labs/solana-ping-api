package main

import (
	"context"
	"log"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	influxdb2write "github.com/influxdata/influxdb-client-go/v2/api/write"
)

// InfluxdbAsyncBatchSize influxdb asynchronize batch size
const InfluxdbAsyncBatchSize = 20

// InfluxdbClient is a client to connect to influxdb in influx cloud
type InfluxdbClient struct {
	Bucket         string
	Organization   string
	AccessToken    string
	InfluxCloudURL string
	Client         influxdb2.Client
}

// NewInfluxdbClient Create a new InfluxClient
func NewInfluxdbClient(config InfluxdbConfig) *InfluxdbClient {
	c := new(InfluxdbClient)
	c.Bucket = config.Bucket
	c.AccessToken = config.AccessToken
	c.InfluxCloudURL = config.InfluxdbURL
	c.Client = influxdb2.NewClientWithOptions(c.InfluxCloudURL, c.AccessToken,
		influxdb2.DefaultOptions().SetBatchSize(InfluxdbAsyncBatchSize))
	return c
}

// PrepareInfluxdbData prepare influxdb datapoint form PingResult
func (i *InfluxdbClient) PrepareInfluxdbData(r PingResult) *influxdb2write.Point {
	return influxdb2.NewPoint(r.Cluster,
		map[string]string{},
		map[string]interface{}{
			"hostname":             r.Hostname,
			"compute_unit_price":   r.ComputeUnitPrice,
			"request_compute_unit": r.RequestComputeUnits,
			"submit":               r.Submitted,
			"confirmed":            r.Confirmed,
			"loss":                 r.Loss,
			"max":                  r.Max,
			"min":                  r.Min,
			"mean":                 r.Mean,
			"stddev":               r.Stddev,
			"take_time":            r.TakeTime,
			"error":                r.Error,
		},
		time.Now())

}

// SendDatapointAsync send datapoint to influx cloud
func (i *InfluxdbClient) SendDatapointAsync(p *influxdb2write.Point) {
	if i.Client == nil {
		log.Println("ERROR! InfluxClient has not initiated!")
		return
	}
	go func() {
		writeAPI := i.Client.WriteAPIBlocking(i.Organization, i.Bucket)
		err := writeAPI.WritePoint(context.Background(), p)
		if err != nil {
			log.Println("Influxdb write ppint ERROR!")
		}
		writeAPI.Flush(context.Background())
	}()
}

// ClientClose close Client connection
func (i *InfluxdbClient) ClientClose() {
	i.Client.Close()
}
