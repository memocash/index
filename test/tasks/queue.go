package tasks

import (
	"bytes"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/test/run/queue"
	"github.com/memocash/index/test/suite"
)

const (
	topic = "test"
	shard = 0
)

var (
	bytesAB = []byte("ab")
	itemAA  = queue.Item{
		Topic: topic,
		Uid:   []byte("aa"),
		Data:  jutil.GetUint32Data(1),
	}
	itemAB = queue.Item{
		Topic: topic,
		Uid:   bytesAB,
		Data:  jutil.GetUint32Data(2),
	}
	itemBA = queue.Item{
		Topic: topic,
		Uid:   []byte("ba"),
		Data:  jutil.GetUint32Data(3),
	}
	itemCA = queue.Item{
		Topic: topic,
		Uid:   []byte("ca"),
		Data:  jutil.GetUint32Data(4),
	}
	itemCB = queue.Item{
		Topic: topic,
		Uid:   []byte("cb"),
		Data:  jutil.GetUint32Data(5),
	}
)

func checkExpectedItems(items, expectedItems []queue.Item) error {
	if len(items) != len(expectedItems) {
		return jerr.Newf("error expected %d items, got %d", len(expectedItems), len(items))
	}
	for i, item := range items {
		if !bytes.Equal(item.Uid, expectedItems[i].Uid) {
			return jerr.Newf("error item %d uid (%x) doesn't match expected (%x)",
				i, item.Uid, expectedItems[i].Uid)
		}
	}
	return nil
}

var queueTest = suite.Test{
	Name: TestQueue,
	Test: func(r *suite.TestRequest) error {
		var itemList = []queue.Item{
			itemAA,
			itemAB,
			itemBA,
			itemCA,
			itemCB,
		}
		var expectedItems = []queue.Item{
			itemAA,
			itemAB,
			itemCA,
			itemCB,
		}
		add := queue.NewAdd(shard)
		if err := add.Add(itemList); err != nil {
			return jerr.Get("error adding items to queue", err)
		}
		get := queue.NewGet(shard)
		if err := get.GetByPrefixes(topic, [][]byte{[]byte("a"), []byte("c")}); err != nil {
			return jerr.Get("error getting by prefix", err)
		}
		if err := checkExpectedItems(get.Items, expectedItems); err != nil {
			return jerr.Get("error checking expected items", err)
		}
		return nil
	},
}

var waitTest = suite.Test{
	Name: TestQueueWait,
	Test: func(r *suite.TestRequest) error {
		add := queue.NewAdd(shard)
		var itemList1 = []queue.Item{
			itemAA,
			itemAB,
		}
		if err := add.Add(itemList1); err != nil {
			return jerr.Get("error adding items 1 to queue", err)
		}
		get := queue.NewGet(shard)
		if err := get.GetAndWait(topic, nil); err != nil {
			return jerr.Get("error getting and waiting 1", err)
		}
		if err := checkExpectedItems(get.Items, itemList1); err != nil {
			return jerr.Get("error checking expected items 1", err)
		}
		var response = make(chan error)
		go func() {
			err := get.GetAndWait(topic, client.IncrementBytes(bytesAB))
			if err != nil {
				response <- jerr.Get("error getting and waiting 2", err)
			} else {
				response <- nil
			}
		}()
		var itemList2 = []queue.Item{
			itemBA,
			itemCA,
			itemCB,
		}
		if err := add.Add(itemList2); err != nil {
			return jerr.Get("error adding items 2 to queue", err)
		}
		if err := <-response; err != nil {
			return jerr.Get("error with wait response", err)
		}
		if err := checkExpectedItems(get.Items, itemList2); err != nil {
			return jerr.Get("error checking expected items 2", err)
		}
		return nil
	},
}
