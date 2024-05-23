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

type Disconnect struct {
	Message string
}

func (i *Disconnect) Disconnect(nodeId string) error {
	jsonData, err := json.Marshal(admin.NodeDisconnectRequest{
		NodeId: nodeId,
	})
	if err != nil {
		return fmt.Errorf("error marshalling data; %w", err)
	}
	url := "http://" + config.GetHost(config.GetAdminPort()) + admin.UrlNodeDisconnect
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error getting node disconnect; %w", err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading node disconnect body; %w", err)
	}
	i.Message = string(body)
	return nil
}

func NewDisconnect() *Disconnect {
	return &Disconnect{}
}
