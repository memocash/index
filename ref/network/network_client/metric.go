package network_client

import (
	"context"
	"encoding/hex"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/server/ref/network/gen/network_pb"
	"google.golang.org/grpc"
	"time"
)

const (
	MetricIdSize = 32
)

func GetEmptyParent() [MetricIdSize]byte {
	return [MetricIdSize]byte{}
}

func GetEmptyParentAny() []byte {
	parent := GetEmptyParent()
	return parent[:]
}

type MetricInfo struct {
	Id       []byte
	Parent   []byte
	Action   string
	Order    int32
	Count    int32
	Start    time.Time
	Duration time.Duration
	Children []*MetricInfo
}

func (i MetricInfo) IdString() string {
	return hex.EncodeToString(i.Id)
}

func (i MetricInfo) IdShort() string {
	return jutil.ShortHash(i.IdString())
}

func (i MetricInfo) ParentString() string {
	return hex.EncodeToString(i.Parent)
}

func (i MetricInfo) ParentShort() string {
	return jutil.ShortHash(i.ParentString())
}

func (i *MetricInfo) SetChildren(tree bool) error {
	metricGetter := NewMetrics()
	err := metricGetter.GetByParentId(i.Id)
	if err != nil {
		return jerr.Get("error getting by parent id", err)
	}
	i.Children = metricGetter.Infos
	if tree {
		for _, child := range i.Children {
			err := child.SetChildren(tree)
			if err != nil {
				return jerr.Getf(err, "error setting children of children: %x", child.Id)
			}
		}
	}
	return nil
}

type Metrics struct {
	Infos []*MetricInfo
}

func (m *Metrics) GetById(id []byte) error {
	err := m.Get(id, nil)
	if err != nil {
		return jerr.Get("error getting metric by id", err)
	}
	return nil
}

func (m *Metrics) GetByParentId(parentId []byte) error {
	err := m.Get(nil, [][]byte{parentId})
	if err != nil {
		return jerr.Get("error getting metrics by parent id", err)
	}
	return nil
}

func (m *Metrics) GetTree(parentId []byte) error {
	err := m.GetById(parentId)
	if err != nil {
		return jerr.Get("error getting parent", err)
	}
	if len(m.Infos) != 1 {
		return jerr.Newf("error unexpected number of infos: %d", len(m.Infos))
	}
	err = m.Infos[0].SetChildren(true)
	if err != nil {
		return jerr.Get("error setting info children", err)
	}
	return nil
}

func (m *Metrics) Get(id []byte, parentIds [][]byte) error {
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
	metricList, err := c.GetMetrics(ctx, &network_pb.MetricRequest{
		Id:      id,
		Parents: parentIds,
	})
	if err != nil {
		return jerr.Get("error getting rpc network balance by address", err)
	}
	for _, info := range metricList.Infos {
		m.Infos = append(m.Infos, &MetricInfo{
			Id:       info.Id,
			Parent:   info.Parent,
			Action:   info.Action,
			Order:    info.Order,
			Count:    info.Count,
			Start:    time.Unix(info.Start, 0),
			Duration: time.Duration(info.Duration),
		})
	}
	return nil
}

func NewMetrics() *Metrics {
	return &Metrics{
	}
}
