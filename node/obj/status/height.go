package status

import (
	"fmt"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item"
)

const NoShard = -1

const (
	NameBlock       = "block"
	NameUtxo        = "utxo"
	NameDoubleSpend = "double-spend"
	NameLockHeight  = "lock-height"
	NameMemo        = "memo"
)

type BlockHeight struct {
	Height int64
	Block  []byte
}

type Height struct {
	Name        string
	StartHeight int64
}

func (s *Height) error(err error) {
	jerr.Get("error saving tx queue", err).Print()
}

func (s *Height) SetHeight(blockHeight BlockHeight) error {
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

func (s *Height) GetHeight() (BlockHeight, error) {
	if s.StartHeight != 0 {
		if s.StartHeight == -1 {
			return BlockHeight{}, nil
		}
		return BlockHeight{
			Height: s.StartHeight,
		}, nil
	}
	heightProcessed, err := item.GetRecentHeightProcessed(s.Name)
	if err != nil {
		if client.IsMessageNotSetError(err) {
			return BlockHeight{}, nil
		}
		return BlockHeight{}, jerr.Get("error getting max height processed", err)
	}
	return BlockHeight{
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

func GetStatusShardName(name string, shard int) string {
	if shard == NoShard {
		return name
	}
	return fmt.Sprintf("%s-%d", name, shard)
}
