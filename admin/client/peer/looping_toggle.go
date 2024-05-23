package peer

import (
	"fmt"
	"github.com/memocash/index/admin/admin"
	"github.com/memocash/index/ref/config"
	"io/ioutil"
	"net/http"
)

type LoopingToggle struct {
	Message string
}

func (i *LoopingToggle) Enable() error {
	if err := i.set(admin.UrlNodeLoopingEnable); err != nil {
		return fmt.Errorf("error enabling looping; %w", err)
	}
	return nil
}

func (i *LoopingToggle) Disable() error {
	if err := i.set(admin.UrlNodeLoopingDisable); err != nil {
		return fmt.Errorf("error disabling looping; %w", err)
	}
	return nil
}

func (i *LoopingToggle) set(url string) error {
	resp, err := http.Get("http://" + config.GetHost(config.GetAdminPort()) + url)
	if err != nil {
		return fmt.Errorf("error getting node looping toggle; %w", err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading node looping toggle body; %w", err)
	}
	i.Message = string(body)
	return nil
}

func NewLoopingToggle() *LoopingToggle {
	return &LoopingToggle{}
}
