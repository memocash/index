package peer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/memocash/index/admin/admin"
	"github.com/memocash/index/ref/config"
	"io/ioutil"
	"net/http"
)

type History struct {
	Connections []admin.Connection
}

func (c *History) Get() error {
	jsonData, err := json.Marshal(admin.NodeHistoryRequest{
		SuccessOnly: true,
	})
	if err != nil {
		return fmt.Errorf("error marshalling history request data; %w", err)
	}
	url := "http://" + config.GetHost(config.GetAdminPort()) + admin.UrlNodeHistory
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error getting peer node history; %w", err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading peer list history; %w", err)
	}
	var historyResponse = new(admin.NodeHistoryResponse)
	if err := json.Unmarshal(body, historyResponse); err != nil {
		return fmt.Errorf("error unmarshalling node history response; %w", err)
	}
	c.Connections = historyResponse.Connections
	return nil
}

func NewHistory() *History {
	return &History{}
}
