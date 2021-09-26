package peer

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/admin"
	"github.com/memocash/server/ref/config"
	"io/ioutil"
	"net/http"
)

type ConnectNext struct {
	Message string
}

func (i *ConnectNext) Get() error {
	resp, err := http.Get("http://" + config.GetHost(config.GetAdminPort()) + admin.UrlNodeConnectNext)
	if err != nil {
		return jerr.Get("error getting node connect next", err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return jerr.Get("error reading node connect next body", err)
	}
	i.Message = string(body)
	return nil
}

func NewConnectNext() *ConnectNext {
	return &ConnectNext{}
}
