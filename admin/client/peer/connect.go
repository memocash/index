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

type Connect struct {
	Message string
}

func (i *Connect) Connect(ip []byte, port uint16) error {
	jsonData, err := json.Marshal(admin.NodeConnectRequest{
		Ip:   ip,
		Port: port,
	})
	if err != nil {
		return fmt.Errorf("error marshalling connect request data; %w", err)
	}
	url := "http://" + config.GetHost(config.GetAdminPort()) + admin.UrlNodeConnect
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error getting node connect; %w", err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading node connect default body; %w", err)
	}
	i.Message = string(body)
	return nil
}

func NewConnect() *Connect {
	return &Connect{}
}
