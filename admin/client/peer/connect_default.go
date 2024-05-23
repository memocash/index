package peer

import (
	"fmt"
	"github.com/memocash/index/admin/admin"
	"github.com/memocash/index/ref/config"
	"io/ioutil"
	"net/http"
)

type ConnectDefault struct {
	Message string
}

func (i *ConnectDefault) Get() error {
	resp, err := http.Get("http://" + config.GetHost(config.GetAdminPort()) + admin.UrlNodeConnectDefault)
	if err != nil {
		return fmt.Errorf("error getting node connect default; %w", err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading node connect default body; %w", err)
	}
	i.Message = string(body)
	return nil
}

func NewConnectDefault() *ConnectDefault {
	return &ConnectDefault{}
}
