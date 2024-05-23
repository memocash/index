package peer

import (
	"bytes"
	"encoding/json"
	"fmt"
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
		return fmt.Errorf("error marshalling found peers request data; %w", err)
	}
	url := "http://" + config.GetHost(config.GetAdminPort()) + admin.UrlNodeFoundPeers
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error getting found peers; %w", err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading found peers body; %w", err)
	}
	var foundPeersResponse = new(admin.NodeFoundPeersResponse)
	if err := json.Unmarshal(body, foundPeersResponse); err != nil {
		return fmt.Errorf("error unmarshalling node history response; %w", err)
	}
	c.FoundPeers = foundPeersResponse.FoundPeers
	return nil
}

func NewFoundPeers() *FoundPeers {
	return &FoundPeers{}
}
