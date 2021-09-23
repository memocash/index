package client

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/admin"
	"github.com/memocash/server/ref/config"
	"io/ioutil"
	"net/http"
)

type PeerConnect struct {
	Message string
}

func (i *PeerConnect) Get() error {
	resp, err := http.Get("http://" + config.GetHost(config.GetAdminPort()) + admin.UrlNodeConnectDefault)
	if err != nil {
		return jerr.Get("error getting node connect default", err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return jerr.Get("error reading node connect default body", err)
	}
	i.Message = string(body)
	return nil
}

func NewPeerConnect() *PeerConnect {
	return &PeerConnect{}
}
