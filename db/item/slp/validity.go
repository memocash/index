package slp

import (
	"context"
	"fmt"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/memo"
)

const (
	ValidityStatusValid   = 1
	ValidityStatusInvalid = 2
)

// Validity is a per-tx SLP validity verdict. Absence of a row means the tx is
// still pending (undecided). A written verdict is final (DAG-validity).
type Validity struct {
	TxHash [32]byte
	Status uint8
	Reason uint8
}

func (v *Validity) GetTopic() string {
	return db.TopicSlpValidity
}

func (v *Validity) GetShardSource() uint {
	return client.GenShardSource(v.TxHash[:])
}

func (v *Validity) GetUid() []byte {
	return jutil.ByteReverse(v.TxHash[:])
}

func (v *Validity) SetUid(uid []byte) {
	if len(uid) != memo.TxHashLength {
		return
	}
	copy(v.TxHash[:], jutil.ByteReverse(uid))
}

func (v *Validity) Serialize() []byte {
	return []byte{v.Status, v.Reason}
}

func (v *Validity) Deserialize(data []byte) {
	if len(data) != 2 {
		return
	}
	v.Status = data[0]
	v.Reason = data[1]
}

func (v *Validity) IsValid() bool {
	return v.Status == ValidityStatusValid
}

func GetValidities(ctx context.Context, txHashes [][32]byte) ([]*Validity, error) {
	var shardUids = make(map[uint32][][]byte)
	for _, txHash := range txHashes {
		shard := db.GetShardIdFromByte32(txHash[:])
		shardUids[shard] = append(shardUids[shard], jutil.ByteReverse(txHash[:]))
	}
	messages, err := db.GetSpecific(ctx, db.TopicSlpValidity, shardUids)
	if err != nil {
		return nil, fmt.Errorf("error getting slp validities; %w", err)
	}
	var validities []*Validity
	for i := range messages {
		var validity = new(Validity)
		db.Set(validity, messages[i])
		validities = append(validities, validity)
	}
	return validities, nil
}
