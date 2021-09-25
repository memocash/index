package peer

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/admin"
	"github.com/memocash/server/ref/config"
	"io/ioutil"
	"net/http"
)

type History struct {
	Connections string
}

func (c *History) Get() error {
	resp, err := http.Get("http://" + config.GetHost(config.GetAdminPort()) + admin.UrlNodeHistory)
	if err != nil {
		return jerr.Get("error getting peer node history", err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return jerr.Get("error reading peer list history", err)
	}
	c.Connections = string(body)
	return nil
}

func NewHistory() *History {
	return &History{}
}
