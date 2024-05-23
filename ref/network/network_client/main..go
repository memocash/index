package network_client

import (
	"context"
	"fmt"
	"github.com/memocash/index/ref/config"
	"github.com/memocash/index/ref/network/gen/network_pb"
	"google.golang.org/grpc"
	"time"
)

var _config config.RpcConfig

func SetConfig(config config.RpcConfig) {
	_config = config
}

func GetConfig() config.RpcConfig {
	return _config
}

type Connection struct {
	Client network_pb.NetworkClient
	conn   *grpc.ClientConn
	cancel context.CancelFunc
}

func (c *Connection) connect() error {
	cfg := GetConfig()
	if !cfg.IsSet() {
		return fmt.Errorf("error network client config not set; %w", config.NotSetError)
	}
	conn, err := grpc.Dial(cfg.String(), grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("error did not connect network; %w", err)
	}
	c.conn = conn
	c.Client = network_pb.NewNetworkClient(c.conn)
	return nil
}

func (c *Connection) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
	c.cancel()
}

func (c *Connection) GetTimeoutContext(d time.Duration) context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), d)
	c.cancel = cancel
	return ctx
}

func (c *Connection) GetDefaultContext() context.Context {
	return c.GetTimeoutContext(time.Hour)
}

func NewConnection() (*Connection, error) {
	var conn = new(Connection)
	if err := conn.connect(); err != nil {
		return nil, fmt.Errorf("error connecting; %w", err)
	}
	return conn, nil
}
