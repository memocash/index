package peer

import (
	"fmt"
	"github.com/memocash/index/admin/admin"
	"github.com/memocash/index/ref/config"
	"io/ioutil"
	"net/http"
)

type ListConnections struct {
	Connections string
}

func (c *ListConnections) List() error {
	resp, err := http.Get("http://" + config.GetHost(config.GetAdminPort()) + admin.UrlNodeListConnections)
	if err != nil {
		return fmt.Errorf("error getting peer list connections; %w", err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading peer list connections body; %w", err)
	}
	c.Connections = string(body)
	return nil
}

func NewListConnections() *ListConnections {
	return &ListConnections{}
}
