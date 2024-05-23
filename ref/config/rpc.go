package config

import (
	"fmt"
)

type RpcConfig struct {
	Host string
	Port int
}

func (r RpcConfig) String() string {
	return fmt.Sprintf("[%s]:%d", r.Host, r.Port)
}

func (r RpcConfig) IsSet() bool {
	return r.Host != "" && r.Port != 0
}

var NotSetError = fmt.Errorf("error rpc config is not set")
