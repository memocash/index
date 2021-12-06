package peer

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/admin/admin"
	"github.com/memocash/index/ref/config"
	"io/ioutil"
	"net/http"
)

type Get struct {
	Message string
}

func (i *Get) Get() error {
	resp, err := http.Get("http://" + config.GetHost(config.GetAdminPort()) + admin.UrlNodeGetAddrs)
	if err != nil {
		return jerr.Get("error getting admin index", err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return jerr.Get("error reading body", err)
	}
	i.Message = string(body)
	return nil
}

func NewGet() *Get {
	return &Get{}
}
