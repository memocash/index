package chain

import "github.com/memocash/index/db/item/db"

func GetTopics() []db.Object {
	return []db.Object{
		&Block{},
		&BlockHeight{},
		&BlockInfo{},
		&BlockTx{},
		&HeightBlock{},
		&HeightDuplicate{},
		&OutputInput{},
		&OutputInputSingle{},
		&Tx{},
		&TxBlock{},
		&TxInput{},
		&TxOutput{},
		&TxProcessed{},
		&TxSeen{},
	}
}
