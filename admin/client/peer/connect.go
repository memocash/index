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

type Connect struct {
	Message string
}

func (i *Connect) Connect(ip []byte, port uint16) error {
	jsonData, err := json.Marshal(admin.NodeConnectRequest{
		Ip:   ip,
		Port: port,
	})
	if err != nil {
		return jerr.Get("error marshalling connect request data", err)
	}
	url := "http://" + config.GetHost(config.GetAdminPort()) + admin.UrlNodeConnect
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return jerr.Get("error getting node connect", err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return jerr.Get("error reading node connect default body", err)
	}
	i.Message = string(body)
	return nil
}

func NewConnect() *Connect {
	return &Connect{}
}
