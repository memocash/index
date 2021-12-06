package network_client

import (
	"context"
	"encoding/hex"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/ref/network/gen/network_pb"
	"google.golang.org/grpc"
	"time"
)

type MetricTime struct {
	Id   []byte
	Time time.Time
}

func (t MetricTime) IdString() string {
	return hex.EncodeToString(t.Id)
}

func (t MetricTime) IdShort() string {
	return jutil.ShortHash(t.IdString())
}

type MetricTimeGetter struct {
	MetricTimes []*MetricTime
}

func (g *MetricTimeGetter) Get(start time.Time) error {
	rpcConfig := GetConfig()
	if !rpcConfig.IsSet() {
		return jerr.New("error config not set")
	}
	conn, err := grpc.Dial(rpcConfig.String(), grpc.WithInsecure())
	if err != nil {
		return jerr.Get("error dial grpc did not connect network", err)
	}
	defer conn.Close()
	c := network_pb.NewNetworkClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
	defer cancel()
	metricList, err := c.GetMetricList(ctx, &network_pb.MetricTimeRequest{
		Start: start.Unix(),
	})
	if err != nil {
		return jerr.Get("error getting rpc network balance by address", err)
	}
	for _, metric := range metricList.Metrics {
		g.MetricTimes = append(g.MetricTimes, &MetricTime{
			Id:   metric.Id,
			Time: time.Unix(metric.Time, 0),
		})
	}
	return nil
}

func NewMetricTimeGetter() *MetricTimeGetter {
	return &MetricTimeGetter{}
}
