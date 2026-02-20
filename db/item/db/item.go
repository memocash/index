package db

import (
	"context"
	"fmt"
	"sync"

	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/ref/bitcoin/wallet"
	"github.com/memocash/index/ref/config"
)

func GetItem(ctx context.Context, obj Object) error {
	shardConfig := config.GetShardConfig(uint32(obj.GetShardSource()), config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	if err := dbClient.GetSingle(ctx, obj.GetTopic(), obj.GetUid()); err != nil && !client.IsMessageNotSetError(err) {
		return fmt.Errorf("error getting db item single; %w", err)
	}
	if len(dbClient.Messages) != 1 {
		return fmt.Errorf("error item not found; %w", client.EntryNotFoundError)
	}
	obj.Deserialize(dbClient.Messages[0].Message)
	return nil
}

func shardReversedTxHashes(txHashes [][32]byte) map[uint32][][]byte {
	var shardUids = make(map[uint32][][]byte)
	for _, txHash := range txHashes {
		shard := GetShardIdFromByte32(txHash[:])
		shardUids[shard] = append(shardUids[shard], jutil.ByteReverse(txHash[:]))
	}
	return shardUids
}

func ShardUidsTxHashes(txHashes [][32]byte) map[uint32][][]byte {
	return shardReversedTxHashes(txHashes)
}

func ShardUidsAddrs(addrs [][25]byte) map[uint32][][]byte {
	return ShardUids(wallet.AddrsToSlices(addrs))
}

func ShardUids(byteUids [][]byte) map[uint32][][]byte {
	var shardUids = make(map[uint32][][]byte)
	for _, uid := range byteUids {
		shard := GetShardIdFromByte32(uid)
		shardUids[shard] = append(shardUids[shard], uid)
	}
	return shardUids
}

func GetSpecific(ctx context.Context, topic string, shardUids map[uint32][][]byte) ([]client.Message, error) {
	wait := NewWait(len(shardUids))
	var messages []client.Message
	for shardT, uidsT := range shardUids {
		go func(shard uint32, uids [][]byte) {
			defer wait.Group.Done()
			uids = jutil.RemoveDupesAndEmpties(uids)
			dbClient := GetShardClient(shard)
			for len(uids) > 0 {
				var uidsToUse [][]byte
				if len(uids) > client.HugeLimit {
					uidsToUse, uids = uids[:client.HugeLimit], uids[client.HugeLimit:]
				} else {
					uidsToUse, uids = uids, nil
				}
				if err := dbClient.GetSpecific(ctx, topic, uidsToUse); err != nil {
					wait.AddError(fmt.Errorf("error getting client message get specific; %w", err))
					return
				}
				wait.Lock.Lock()
				messages = append(messages, dbClient.Messages...)
				wait.Lock.Unlock()
			}
		}(shardT, uidsT)
	}
	wait.Group.Wait()
	if len(wait.Errs) > 0 {
		return nil, fmt.Errorf("error getting specific messages; %w", jerr.Combine(wait.Errs...))
	}
	return messages, nil
}

func removeDupeAndEmptyPrefixes(prefixes []client.Prefix) []client.Prefix {
	var seen = make(map[string]bool)
	var newPrefixes []client.Prefix
	for _, prefix := range prefixes {
		if len(prefix.Prefix) == 0 {
			continue
		}
		if _, ok := seen[string(prefix.Prefix)]; ok {
			continue
		}
		seen[string(prefix.Prefix)] = true
		newPrefixes = append(newPrefixes, prefix)
	}
	return newPrefixes
}

func ShardPrefixesTxHashes(txHashes [][32]byte) map[uint32][]client.Prefix {
	var shardPrefixes = make(map[uint32][]client.Prefix)
	for shard, uids := range shardReversedTxHashes(txHashes) {
		for _, uid := range uids {
			shardPrefixes[shard] = append(shardPrefixes[shard], client.NewPrefix(uid))
		}
	}
	return shardPrefixes
}

func ShardPrefixesAddrs(addrs [][25]byte) map[uint32][]client.Prefix {
	return ShardPrefixes(wallet.AddrsToSlices(addrs))
}

func ShardPrefixes(bytePrefixes [][]byte) map[uint32][]client.Prefix {
	var shardPrefixes = make(map[uint32][]client.Prefix)
	for _, bytePrefix := range bytePrefixes {
		shard := GetShardIdFromByte32(bytePrefix)
		shardPrefixes[shard] = append(shardPrefixes[shard], client.NewPrefix(bytePrefix))
	}
	return shardPrefixes
}

func GetByPrefixes(ctx context.Context, topic string, shardPrefixes map[uint32][]client.Prefix, opts ...client.Option) ([]client.Message, error) {
	wait := NewWait(len(shardPrefixes))
	var messages []client.Message
	for shardT, prefixesT := range shardPrefixes {
		go func(shard uint32, prefixes []client.Prefix) {
			defer wait.Group.Done()
			prefixes = removeDupeAndEmptyPrefixes(prefixes)
			dbClient := GetShardClient(shard)
			for len(prefixes) > 0 {
				var prefixesToUse []client.Prefix
				if len(prefixes) > client.HugeLimit {
					prefixesToUse, prefixes = prefixes[:client.HugeLimit], prefixes[client.HugeLimit:]
				} else {
					prefixesToUse, prefixes = prefixes, nil
				}
				if err := dbClient.GetByPrefixes(ctx, topic, prefixesToUse, opts...); err != nil {
					wait.AddError(fmt.Errorf("error getting client message get by prefixes; %w", err))
					return
				}
				wait.Lock.Lock()
				messages = append(messages, dbClient.Messages...)
				wait.Lock.Unlock()
			}
		}(shardT, prefixesT)
	}
	wait.Group.Wait()
	if len(wait.Errs) > 0 {
		return nil, fmt.Errorf("error getting prefix messages; %w", jerr.Combine(wait.Errs...))
	}
	return messages, nil
}

func ListenPrefixes(ctx context.Context, topic string, shardPrefixes map[uint32][][]byte) (chan *client.Message, error) {
	var chanMessages = make(chan *client.Message)
	var once sync.Once
	for shard, prefixes := range shardPrefixes {
		shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
		chanMessage, err := client.NewClient(shardConfig.GetHost()).Listen(ctx, topic, prefixes)
		if err != nil {
			return nil, fmt.Errorf("error getting listen messages chan shard: %d; %w", shard, err)
		}
		go func() {
			defer once.Do(func() {
				close(chanMessages)
			})
			for {
				select {
				case <-ctx.Done():
					return
				case msg, ok := <-chanMessage:
					if !ok {
						return
					}
					chanMessages <- msg
				}
			}
		}()
	}
	return chanMessages, nil
}
