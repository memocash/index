package peer

import (
	"bytes"
	"encoding/json"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/admin/admin"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/ref/config"
	"io/ioutil"
	"net/http"
)

type FoundPeers struct {
	FoundPeers []*item.FoundPeer
}

func (c *FoundPeers) Get(ip []byte, port uint16) error {
	jsonData, err := json.Marshal(admin.NodeFoundPeersRequest{
		Ip:   ip,
		Port: port,
	})
	if err != nil {
		return jerr.Get("error marshalling found peers request data", err)
	}
	url := "http://" + config.GetHost(config.GetAdminPort()) + admin.UrlNodeFoundPeers
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return jerr.Get("error getting found peers", err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return jerr.Get("error reading found peers body", err)
	}
	var foundPeersResponse = new(admin.NodeFoundPeersResponse)
	if err := json.Unmarshal(body, foundPeersResponse); err != nil {
		return jerr.Get("error unmarshalling node history response", err)
	}
	c.FoundPeers = foundPeersResponse.FoundPeers
	return nil
}

func NewFoundPeers() *FoundPeers {
	return &FoundPeers{}
}
