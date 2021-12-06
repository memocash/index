package parse

import (
	"github.com/jchavannes/btcd/txscript"
	"github.com/jchavannes/btcd/wire"
	"github.com/memocash/index/ref/bitcoin/memo"
)

type Meta struct {
	OpReturn   *wire.TxOut
	OutputType memo.OutputType
	Multi      bool
}

func (m Meta) GetSpecialIndexes() []uint {
	var specialIndexes []uint
	switch m.OutputType {
	case memo.OutputTypeTokenCreate, memo.OutputTypeTokenCreateNftGroup, memo.OutputTypeTokenCreateNftChild:
		slpCreateParser := NewSlpCreate()
		if err := slpCreateParser.Parse(m.OpReturn.PkScript); err == nil {
			if slpCreateParser.Quantity > 0 {
				specialIndexes = append(specialIndexes, memo.SlpMintTokenIndex)
			}
			if slpCreateParser.BatonIndex > 0 && slpCreateParser.BatonIndex != memo.SlpNftChildBatonVOut {
				specialIndexes = append(specialIndexes, uint(slpCreateParser.BatonIndex))
			}
		}
	case memo.OutputTypeTokenMint, memo.OutputTypeTokenMintNftGroup:
		slpMintParser := NewSlpMint()
		if err := slpMintParser.Parse(m.OpReturn.PkScript); err == nil {
			if slpMintParser.Quantity > 0 {
				specialIndexes = append(specialIndexes, memo.SlpMintTokenIndex)
			}
			if slpMintParser.BatonIndex > 0 {
				specialIndexes = append(specialIndexes, uint(slpMintParser.BatonIndex))
			}
		}
	case memo.OutputTypeTokenSend, memo.OutputTypeTokenSendNftGroup, memo.OutputTypeTokenSendNftChild:
		slpSendParser := NewSlpSend()
		if err := slpSendParser.Parse(m.OpReturn.PkScript); err == nil {
			for index, amount := range slpSendParser.Quantities {
				if amount > 0 {
					specialIndexes = append(specialIndexes, uint(index+1))
				}
			}
		}
	}
	return specialIndexes
}

func GetMeta(txMsg *wire.MsgTx) *Meta {
	var meta = new(Meta)
	for _, txOut := range txMsg.TxOut {
		if len(txOut.PkScript) < 5 || txOut.PkScript[0] != txscript.OP_RETURN {
			continue
		}
		if meta.OpReturn != nil {
			meta.Multi = true
			continue
		}
		meta.OpReturn = txOut
		meta.OutputType = memo.GetOutputTypeNew(txOut.PkScript)
	}
	return meta
}
