package peer

import (
	"fmt"
	"github.com/memocash/index/admin/admin"
	"github.com/memocash/index/ref/config"
	"io/ioutil"
	"net/http"
)

type ConnectNext struct {
	Message string
}

func (i *ConnectNext) Get() error {
	resp, err := http.Get("http://" + config.GetHost(config.GetAdminPort()) + admin.UrlNodeConnectNext)
	if err != nil {
		return fmt.Errorf("error getting node connect next; %w", err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading node connect next body; %w", err)
	}
	i.Message = string(body)
	return nil
}

func NewConnectNext() *ConnectNext {
	return &ConnectNext{}
}
