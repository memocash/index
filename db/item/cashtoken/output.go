package cashtoken

import (
	"context"
	"fmt"

	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/cashtoken"
)

// Output is a token-carrying tx output: the decoded CashTokens prefix keyed
// by the outpoint. Unlike SLP there is no validity to track; a token prefix
// is consensus-enforced, so an output that made it into a tx the node
// accepted carries exactly the token its prefix says.
type Output struct {
	TxHash     [32]byte
	Index      uint32
	Category   [32]byte
	HasNft     bool
	Capability cashtoken.Capability
	Commitment []byte
	Amount     uint64
}

func (o *Output) GetTopic() string {
	return db.TopicCashTokenOutput
}

func (o *Output) GetShardSource() uint {
	return client.GenShardSource(o.TxHash[:])
}

func (o *Output) GetUid() []byte {
	return db.GetTxHashIndexUid(o.TxHash[:], o.Index)
}

func (o *Output) SetUid(uid []byte) {
	if len(uid) != memo.TxHashLength+4 {
		return
	}
	copy(o.TxHash[:], jutil.ByteReverse(uid[:32]))
	o.Index = jutil.GetUint32Big(uid[32:])
}

// Token returns the output's token as the decoder's type.
func (o *Output) Token() cashtoken.Token {
	return cashtoken.Token{
		Category:   o.Category,
		HasNft:     o.HasNft,
		Capability: o.Capability,
		Commitment: o.Commitment,
		Amount:     o.Amount,
	}
}

// outputFixedLen is the serialized prefix before the variable commitment:
// category, bitfield, amount.
const outputFixedLen = memo.TxHashLength + 1 + 8

func (o *Output) Serialize() []byte {
	return jutil.CombineBytes(
		o.Category[:],
		[]byte{o.Token().Bitfield()},
		jutil.GetUint64Data(o.Amount),
		o.Commitment,
	)
}

func (o *Output) Deserialize(data []byte) {
	if len(data) < outputFixedLen {
		return
	}
	copy(o.Category[:], data[:32])
	bitfield := data[32]
	o.HasNft = bitfield&0x20 != 0
	o.Capability = cashtoken.Capability(bitfield & 0x0f)
	o.Amount = jutil.GetUint64(data[33:41])
	o.Commitment = nil
	if len(data) > outputFixedLen {
		o.Commitment = append([]byte{}, data[outputFixedLen:]...)
	}
}

func NewOutput(txHash [32]byte, index uint32, token cashtoken.Token) *Output {
	return &Output{
		TxHash:     txHash,
		Index:      index,
		Category:   token.Category,
		HasNft:     token.HasNft,
		Capability: token.Capability,
		Commitment: token.Commitment,
		Amount:     token.Amount,
	}
}

func GetOutputs(ctx context.Context, outs []memo.Out) ([]*Output, error) {
	var shardUids = make(map[uint32][][]byte)
	for _, out := range outs {
		shard := db.GetShardIdFromByte32(out.TxHash)
		shardUids[shard] = append(shardUids[shard], db.GetTxHashIndexUid(out.TxHash, out.Index))
	}
	messages, err := db.GetSpecific(ctx, db.TopicCashTokenOutput, shardUids)
	if err != nil {
		return nil, fmt.Errorf("error getting cash token outputs; %w", err)
	}
	var outputs []*Output
	for i := range messages {
		var output = new(Output)
		db.Set(output, messages[i])
		outputs = append(outputs, output)
	}
	return outputs, nil
}

func GetOutputsByTxHashes(ctx context.Context, txHashes [][32]byte) ([]*Output, error) {
	messages, err := db.GetByPrefixes(ctx, db.TopicCashTokenOutput, db.ShardPrefixesTxHashes(txHashes))
	if err != nil {
		return nil, fmt.Errorf("error getting cash token outputs by tx hashes; %w", err)
	}
	var outputs []*Output
	for _, msg := range messages {
		var output = new(Output)
		db.Set(output, msg)
		outputs = append(outputs, output)
	}
	return outputs, nil
}
