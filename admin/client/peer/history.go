package peer

import (
	"bytes"
	"encoding/json"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/admin/admin"
	"github.com/memocash/server/db/item"
	"github.com/memocash/server/ref/config"
	"io/ioutil"
	"net/http"
)

type History struct {
	Connections []*item.PeerConnection
}

func (c *History) Get() error {
	jsonData, err := json.Marshal(admin.NodeHistoryRequest{
		SuccessOnly: true,
	})
	if err != nil {
		return jerr.Get("error marshalling history request data", err)
	}
	url := "http://" + config.GetHost(config.GetAdminPort()) + admin.UrlNodeHistory
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return jerr.Get("error getting peer node history", err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return jerr.Get("error reading peer list history", err)
	}
	var historyResponse = new(admin.NodeHistoryResponse)
	if err := json.Unmarshal(body, historyResponse); err != nil {
		return jerr.Get("error unmarshalling node history response", err)
	}
	c.Connections = historyResponse.Connections
	return nil
}

func NewHistory() *History {
	return &History{}
}
