package broadcast_client

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/ref/broadcast/gen/broadcast_pb"
	"github.com/memocash/index/ref/config"
	"google.golang.org/grpc"
)

var _config config.RpcConfig

func SetConfig(config config.RpcConfig) {
	_config = config
}

func GetConfig() config.RpcConfig {
	return _config
}

type Connection struct {
	Client broadcast_pb.BroadcastClient
	conn   *grpc.ClientConn
}

func (c *Connection) connect() error {
	cfg := GetConfig()
	if !cfg.IsSet() {
		return jerr.Get("error broadcast client config not set", config.GetConfigNotSetError())
	}
	conn, err := grpc.Dial(cfg.String(), grpc.WithInsecure())
	if err != nil {
		return jerr.Get("error did not connect broadcast", err)
	}
	c.conn = conn
	c.Client = broadcast_pb.NewBroadcastClient(c.conn)
	return nil
}

func (c *Connection) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

func NewConnection() (*Connection, error) {
	var conn = new(Connection)
	if err := conn.connect(); err != nil {
		return nil, jerr.Get("error connecting broadcast client", err)
	}
	return conn, nil
}
