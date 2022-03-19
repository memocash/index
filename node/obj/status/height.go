package status

import (
	"fmt"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/node/obj/process"
	"strings"
)

const (
	NameBlock       = "block"
	NameUtxo        = "utxo"
	NameDoubleSpend = "double-spend"
	NameLockHeight  = "lock-height"
)

type Height struct {
	Name        string
	StartHeight int64
}

func (s *Height) error(err error) {
	jerr.Get("error saving tx queue", err).Print()
}

func (s *Height) SetHeight(blockHeight process.BlockHeight) error {
	var heightProcessed = &item.HeightProcessed{
		Name:   s.Name,
		Height: blockHeight.Height,
		Block:  blockHeight.Block,
	}
	err := heightProcessed.Save()
	if err != nil {
		return jerr.Get("error saving block height", err)
	}
	return nil
}

func (s *Height) GetHeight() (process.BlockHeight, error) {
	if s.StartHeight != 0 {
		if s.StartHeight == -1 {
			return process.BlockHeight{}, nil
		}
		return process.BlockHeight{
			Height: s.StartHeight,
		}, nil
	}
	heightProcessed, err := item.GetRecentHeightProcessed(s.Name)
	if err != nil {
		if client.IsMessageNotSetError(err) {
			return process.BlockHeight{}, nil
		}
		return process.BlockHeight{}, jerr.Get("error getting max height processed", err)
	}
	return process.BlockHeight{
		Height: heightProcessed.Height,
		Block:  heightProcessed.Block,
	}, nil
}

func NewHeight(name string, startHeight int64) *Height {
	return &Height{
		Name:        name,
		StartHeight: startHeight,
	}
}

func GetStatusShardName(name string, shards []int) string {
	return fmt.Sprintf("%s-%v", name, strings.Trim(strings.Join(strings.Fields(fmt.Sprint(shards)), ","), "[]"))
}
