package metric

import (
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/ref/config"
	"time"
)

const (
	NameTopicSave   = "topic_save"
	NameTopicRead   = "topic_read"
	NameTopicListen = "topic_listen"
	NameListenCount = "listen_count"
)

const (
	FieldQuantity = "quantity"

	TagTopic  = "topic"
	TagSource = "source"
)

type Point struct {
	Measurement string
	Fields      map[string]interface{}
	Tags        map[string]string
}

type InfluxWriter struct {
	Api    api.WriteAPI
	Config config.InfluxConfig
}

func (i *InfluxWriter) Write(p Point) {
	point := influxdb2.NewPoint(p.Measurement, p.Tags, p.Fields, time.Now())
	if i.Config.Source != "" {
		point.AddTag(TagSource, i.Config.Source)
	}
	i.Api.WritePoint(point)
}

var _influxWriter *InfluxWriter

func getInfluxWriter() *InfluxWriter {
	if jutil.IsNil(_influxWriter) {
		influxConfig := config.GetInfluxConfig()
		if influxConfig.Url == "" || influxConfig.Token == "" {
			return nil
		}
		c := influxdb2.NewClient(influxConfig.Url, influxConfig.Token)
		_influxWriter = &InfluxWriter{
			Config: influxConfig,
			Api:    c.WriteAPI(influxConfig.Org, influxConfig.Bucket),
		}
	}
	return _influxWriter
}
