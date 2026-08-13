package store

import (
	"bytes"
	"context"
	"fmt"

	"github.com/memocash/index/db/metric"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

// MaxMessages is the default result cap for reads with no explicit limit.
const MaxMessages = 10 * 1e6 // 10 million

// filterCancelInterval is how many scanned rows pass between context checks;
// the first row is always checked so a dead context fails immediately.
const filterCancelInterval = 4096

// FilterPattern is one arm of a filtered get: uid and data filters with a
// resume cursor and an optional per-arm result cap. A legacy prefix query is
// the special case {Uid: &Pattern{Parts: [][]byte{prefix}, AnchorStart: true}}.
type FilterPattern struct {
	Uid   *Pattern
	Data  *Pattern
	Start []byte // asc: inclusive lower bound; desc: exclusive upper bound
	Max   int
}

type FilterRequest struct {
	Topic    string // required
	Shard    uint   // required
	Patterns []FilterPattern
	Limit    int
	Desc     bool
}

// GetFiltered returns messages matching the request's pattern arms, in arm
// order. Callers page keyset-style: asc resumes with Start = last returned
// uid + 0x00, desc resumes with Start = last returned uid; a short page means
// the range is exhausted. The context stops an in-flight scan: a sparse
// pattern iterates a whole topic, and an abandoned RPC must not keep burning
// disk and CPU.
func GetFiltered(ctx context.Context, request FilterRequest) ([]*Message, error) {
	db, err := getDb(request.Topic, request.Shard)
	if err != nil {
		return nil, fmt.Errorf("error getting db shard %d; %w", request.Shard, err)
	}

	var maxResults = request.Limit
	if maxResults == 0 {
		maxResults = MaxMessages
	}

	var patterns = request.Patterns
	if len(patterns) == 0 {
		patterns = append(patterns, FilterPattern{})
	}

	var messages []*Message
	defer func() {
		metric.AddTopicRead(metric.TopicRead{
			Topic:    request.Topic,
			Quantity: len(messages),
		})
	}()

	for _, pattern := range patterns {
		var patternMax = maxResults - len(messages)
		if pattern.Max > 0 && pattern.Max < patternMax {
			patternMax = pattern.Max
		}
		patternMessages, err := getFilteredMessages(ctx, db, pattern, patternMax, request.Desc)
		if err != nil {
			return nil, fmt.Errorf("error getting filtered messages; %w", err)
		}
		messages = append(messages, patternMessages...)
		if len(messages) >= maxResults {
			break
		}
	}

	return messages, nil
}

// getFilteredMessages scans one arm within the uid pattern's bounded range
// (anchored first part) or the whole topic, returning rows matching both
// patterns, up to maxResults.
func getFilteredMessages(ctx context.Context, db *leveldb.DB, pattern FilterPattern, maxResults int,
	desc bool) ([]*Message, error) {
	var iterRange *util.Range
	if pattern.Uid != nil && pattern.Uid.AnchorStart &&
		len(pattern.Uid.Parts) > 0 && len(pattern.Uid.Parts[0]) > 0 {
		iterRange = util.BytesPrefix(pattern.Uid.Parts[0])
	} else {
		iterRange = &util.Range{}
	}
	if len(pattern.Start) > 0 {
		if desc {
			if iterRange.Limit == nil || bytes.Compare(pattern.Start, iterRange.Limit) < 0 {
				iterRange.Limit = pattern.Start
			}
		} else if bytes.Compare(pattern.Start, iterRange.Start) > 0 {
			iterRange.Start = pattern.Start
		}
	}
	if len(iterRange.Start) == 0 {
		iterRange.Start = nil
	}

	iter := db.NewIterator(iterRange, nil)
	defer iter.Release()

	first, step := iter.First, iter.Next
	if desc {
		first, step = iter.Last, iter.Prev
	}
	var messages []*Message
	var scanned int
	for ok := first(); ok; ok = step() {
		if scanned%filterCancelInterval == 0 {
			if err := ctx.Err(); err != nil {
				return nil, fmt.Errorf("error filter scan canceled; %w", err)
			}
		}
		scanned++
		if pattern.Uid.Match(iter.Key()) && pattern.Data.Match(iter.Value()) {
			messages = append(messages, &Message{
				Uid:     GetPtrSlice(iter.Key()),
				Message: GetPtrSlice(iter.Value()),
			})
			if len(messages) >= maxResults {
				break
			}
		}
	}
	if err := iter.Error(); err != nil {
		return nil, fmt.Errorf("error with filter scan iterator; %w", err)
	}
	return messages, nil
}
