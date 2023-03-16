package db

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/ref/config"
)

func GetItem(obj Object) error {
	shard := obj.GetShard()
	shardConfig := config.GetShardConfig(uint32(shard), config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	if err := dbClient.GetSingle(obj.GetTopic(), obj.GetUid()); err != nil && !client.IsMessageNotSetError(err) {
		return jerr.Get("error getting db item single", err)
	}
	if len(dbClient.Messages) != 1 {
		return jerr.Get("error item not found", client.EntryNotFoundError)
	}
	obj.Deserialize(dbClient.Messages[0].Message)
	return nil
}

func GetSpecific(topic string, shardUids map[uint32][][]byte) ([]client.Message, error) {
	wait := NewWait(len(shardUids))
	var messages []client.Message
	for shardT, uidsT := range shardUids {
		go func(shard uint32, uids [][]byte) {
			defer wait.Group.Done()
			uids = jutil.RemoveDupesAndEmpties(uids)
			shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
			dbClient := client.NewClient(shardConfig.GetHost())
			for len(uids) > 0 {
				var uidsToUse [][]byte
				if len(uids) > client.HugeLimit {
					uidsToUse, uids = uids[:client.HugeLimit], uids[client.HugeLimit:]
				} else {
					uidsToUse, uids = uids, nil
				}
				if err := dbClient.GetSpecific(topic, uidsToUse); err != nil {
					wait.AddError(jerr.Get("error getting client message get specific", err))
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
		return nil, jerr.Get("error getting specific messages", jerr.Combine(wait.Errs...))
	}
	return messages, nil
}

func GetByPrefixes(topic string, shardPrefixes map[uint32][][]byte) ([]client.Message, error) {
	wait := NewWait(len(shardPrefixes))
	var messages []client.Message
	for shardT, prefixesT := range shardPrefixes {
		go func(shard uint32, prefixes [][]byte) {
			defer wait.Group.Done()
			prefixes = jutil.RemoveDupesAndEmpties(prefixes)
			shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
			dbClient := client.NewClient(shardConfig.GetHost())
			for len(prefixes) > 0 {
				var prefixesToUse [][]byte
				if len(prefixes) > client.HugeLimit {
					prefixesToUse, prefixes = prefixes[:client.HugeLimit], prefixes[client.HugeLimit:]
				} else {
					prefixesToUse, prefixes = prefixes, nil
				}
				if err := dbClient.GetByPrefixes(topic, prefixesToUse); err != nil {
					wait.AddError(jerr.Get("error getting client message get by prefixes", err))
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
		return nil, jerr.Get("error getting prefix messages", jerr.Combine(wait.Errs...))
	}
	return messages, nil
}
