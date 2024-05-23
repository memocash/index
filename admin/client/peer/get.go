package peer

import (
	"fmt"
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
		return fmt.Errorf("error getting admin index; %w", err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading body; %w", err)
	}
	i.Message = string(body)
	return nil
}

func NewGet() *Get {
	return &Get{}
}
