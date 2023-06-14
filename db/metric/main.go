package metric

import (
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/ref/config"
)

var influxWriteApi api.WriteAPIBlocking

func getInflux() (api.WriteAPIBlocking, error) {
	if jutil.IsNil(influxWriteApi) {
		influxConfig := config.GetInfluxConfig()
		if influxConfig.Url == "" || influxConfig.Token == "" {
			return nil, nil
		}
		c := influxdb2.NewClient(influxConfig.Url, influxConfig.Token)
		influxWriteApi = c.WriteAPIBlocking(influxConfig.Org, influxConfig.Bucket)
	}
	return influxWriteApi, nil
}
