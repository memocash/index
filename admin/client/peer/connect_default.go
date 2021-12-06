package peer

import (
	"github.com/jchavannes/jgo/jerr"
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
		return jerr.Get("error getting node connect default", err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return jerr.Get("error reading node connect default body", err)
	}
	i.Message = string(body)
	return nil
}

func NewConnectDefault() *ConnectDefault {
	return &ConnectDefault{}
}
