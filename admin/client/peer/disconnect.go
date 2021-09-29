package peer

import (
	"bytes"
	"encoding/json"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/admin/admin"
	"github.com/memocash/server/ref/config"
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
		return jerr.Get("error marshalling data", err)
	}
	url := "http://" + config.GetHost(config.GetAdminPort()) + admin.UrlNodeDisconnect
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return jerr.Get("error getting node disconnect", err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return jerr.Get("error reading node disconnect body", err)
	}
	i.Message = string(body)
	return nil
}

func NewDisconnect() *Disconnect {
	return &Disconnect{}
}
