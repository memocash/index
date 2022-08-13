package item

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
)

type ProcessError struct {
	TxHash []byte
	Error  string
}

func (e ProcessError) GetUid() []byte {
	return jutil.ByteReverse(e.TxHash)
}

func (e ProcessError) GetShard() uint {
	return client.GetByteShard(e.TxHash)
}

func (e ProcessError) GetTopic() string {
	return db.TopicProcessError
}

func (e ProcessError) Serialize() []byte {
	return []byte(e.Error)
}

func (e *ProcessError) SetUid(uid []byte) {
	e.TxHash = jutil.ByteReverse(uid)
}

func (e *ProcessError) Deserialize(data []byte) {
	e.Error = string(data)
}

func LogProcessError(processError *ProcessError) error {
	jlog.Logf("PROCESS ERROR: %s\n", processError.Error)
	if err := db.Save([]db.Object{processError}); err != nil {
		return jerr.Get("error saving process error", err)
	}
	return nil
}
