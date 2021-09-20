package config

import (
	"fmt"
	"github.com/jchavannes/jgo/jerr"
)

type RpcConfig struct {
	Host string
	Port int
}

func (r RpcConfig) String() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

func (r RpcConfig) IsSet() bool {
	return r.Host != "" && r.Port != 0
}

const (
	notSetErrorMessage = "error rpc config is not set"
)

func GetConfigNotSetError() error {
	return jerr.New(notSetErrorMessage)
}

func IsConfigNotSetError(err error) bool {
	return jerr.HasError(err, notSetErrorMessage)
}
