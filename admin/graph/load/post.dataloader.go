package load

import (
	"encoding/hex"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/admin/graph/dataloader"
	"github.com/memocash/index/admin/graph/model"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item"
	"time"
)

var PostLoaderConfig = dataloader.PostLoaderConfig{
	Wait:     2 * time.Millisecond,
	MaxBatch: 100,
	Fetch: func(txHashStrings []string) ([]*model.Post, []error) {
		var posts = make([]*model.Post, len(txHashStrings))
		var errors = make([]error, len(txHashStrings))
		for i, txHashString := range txHashStrings {
			txHash, err := chainhash.NewHashFromStr(txHashString)
			if err != nil {
				errors[i] = jerr.Get("error getting tx hash from string", err)
				continue
			}
			memoPost, err := item.GetMemoPost(txHash.CloneBytes())
			if err != nil && !client.IsEntryNotFoundError(err) {
				errors[i] = jerr.Get("error getting lock memo post", err)
				continue
			}
			if memoPost == nil {
				errors[i] = jerr.New("error post not found: " + txHashString)
				continue
			}
			posts[i] = &model.Post{
				TxHash:   txHash.String(),
				LockHash: hex.EncodeToString(memoPost.LockHash),
				Text:     memoPost.Post,
			}
		}
		return posts, errors
	},
}
