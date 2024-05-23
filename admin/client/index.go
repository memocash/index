package client

import (
	"fmt"
	"github.com/memocash/index/ref/config"
	"io/ioutil"
	"net/http"
)

type Index struct {
	Message string
}

func (i *Index) Get() error {
	resp, err := http.Get("http://" + config.GetHost(config.GetAdminPort()) + "/")
	if err != nil {
		return fmt.Errorf("error getting admin index; %w", err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading body; %w", err)
	}
	i.Message = string(body)
	return nil
}

func NewIndex() *Index {
	return &Index{}
}
