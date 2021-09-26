package peer

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/admin"
	"github.com/memocash/server/ref/config"
	"io/ioutil"
	"net/http"
)

type LoopingToggle struct {
	Message string
}

func (i *LoopingToggle) Enable() error {
	if err := i.set(admin.UrlNodeLoopingEnable); err != nil {
		return jerr.Get("error enabling looping", err)
	}
	return nil
}

func (i *LoopingToggle) Disable() error {
	if err := i.set(admin.UrlNodeLoopingDisable); err != nil {
		return jerr.Get("error disabling looping", err)
	}
	return nil
}

func (i *LoopingToggle) set(url string) error {
	resp, err := http.Get("http://" + config.GetHost(config.GetAdminPort()) + url)
	if err != nil {
		return jerr.Get("error getting node looping toggle", err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return jerr.Get("error reading node looping toggle body", err)
	}
	i.Message = string(body)
	return nil
}

func NewLoopingToggle() *LoopingToggle {
	return &LoopingToggle{}
}
