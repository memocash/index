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
		&Tx{},
		&TxBlock{},
		&TxInput{},
		&TxOutput{},
		&TxProcessed{},
		&TxSeen{},
	}
}
