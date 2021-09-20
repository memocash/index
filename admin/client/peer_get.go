package client

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/admin"
	"github.com/memocash/server/ref/config"
	"io/ioutil"
	"net/http"
)

type PeerGet struct {
	Message string
}

func (i *PeerGet) Get() error {
	resp, err := http.Get("http://" + config.GetHost(config.GetAdminPort()) + admin.UrlNodeGetAddrs)
	if err != nil {
		return jerr.Get("error getting admin index", err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return jerr.Get("error reading body", err)
	}
	i.Message = string(body)
	return nil
}

func NewPeerGet() *PeerGet {
	return &PeerGet{}
}
