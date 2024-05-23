package item

import (
	"fmt"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"log"
)

type ProcessError struct {
	TxHash [32]byte
	Error  string
}

func (e *ProcessError) GetTopic() string {
	return db.TopicProcessError
}

func (e *ProcessError) GetShardSource() uint {
	return client.GenShardSource(e.TxHash[:])
}

func (e *ProcessError) GetUid() []byte {
	return jutil.ByteReverse(e.TxHash[:])
}

func (e *ProcessError) SetUid(uid []byte) {
	copy(e.TxHash[:], jutil.ByteReverse(uid))
}

func (e *ProcessError) Serialize() []byte {
	return []byte(e.Error)
}

func (e *ProcessError) Deserialize(data []byte) {
	e.Error = string(data)
}

func LogProcessError(processError *ProcessError) error {
	log.Printf("PROCESS ERROR (%s): %s\n", chainhash.Hash(processError.TxHash), processError.Error)
	if err := db.Save([]db.Object{processError}); err != nil {
		return fmt.Errorf("error saving process error; %w", err)
	}
	return nil
}
