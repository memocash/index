package op_return

import (
	"bytes"
	"fmt"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/db/item/db"
	dbMemo "github.com/memocash/index/db/item/memo"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/parse"
	"github.com/memocash/index/ref/bitcoin/tx/script"
)

var memoLikeHandler = &Handler{
	prefix: memo.PrefixLike,
	handle: func(info parse.OpReturn) error {
		if len(info.PushData) != 2 {
			if err := item.LogProcessError(&item.ProcessError{
				TxHash: info.TxHash,
				Error:  fmt.Sprintf("invalid set like, incorrect push data (%d)", len(info.PushData)),
			}); err != nil {
				return jerr.Get("error saving process error memo like incorrect push data", err)
			}
			return nil
		}
		postTxHash := info.PushData[1]
		if len(postTxHash) != memo.TxHashLength {
			if err := item.LogProcessError(&item.ProcessError{
				TxHash: info.TxHash,
				Error:  fmt.Sprintf("error like tx hash not correct size: %d", len(postTxHash)),
			}); err != nil {
				return jerr.Get("error saving process error memo like post tx hash", err)
			}
			return nil
		}
		var memoLike = &dbMemo.LockLike{
			LockHash:   info.LockHash,
			Height:     info.Height,
			LikeTxHash: info.TxHash,
			PostTxHash: postTxHash,
		}
		var memoLiked = &dbMemo.Liked{
			PostTxHash: postTxHash,
			Height:     info.Height,
			LikeTxHash: info.TxHash,
			LockHash:   info.LockHash,
		}
		memoPost, err := dbMemo.GetPost(postTxHash)
		if err != nil {
			return jerr.Get("error getting memo post for like op return handler", err)
		}
		var objects = []db.Object{memoLike, memoLiked}
		if memoPost != nil && !bytes.Equal(memoLike.LockHash, memoPost.LockHash) {
			var tip int64
			for _, txOut := range info.Outputs {
				outputLockHash := script.GetLockHash(txOut.PkScript)
				if bytes.Equal(outputLockHash, memoPost.LockHash) {
					tip += txOut.Value
				}
			}
			if tip > 0 {
				objects = append(objects, &dbMemo.LikeTip{
					LikeTxHash: info.TxHash,
					Tip:        tip,
				})
			}
		}
		if err := db.Save(objects); err != nil {
			return jerr.Get("error saving db memo like object", err)
		}
		if info.Height != item.HeightMempool {
			memoLike.Height = item.HeightMempool
			if err := dbMemo.RemoveLockMemoLike(memoLike); err != nil {
				return jerr.Get("error removing db memo like", err)
			}
		}
		return nil
	},
}
