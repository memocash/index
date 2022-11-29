package op_return

import (
	"fmt"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/db/item/db"
	dbMemo "github.com/memocash/index/db/item/memo"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/parse"
	"github.com/memocash/index/ref/bitcoin/wallet"
)

var memoLikeHandler = &Handler{
	prefix: memo.PrefixLike,
	handle: func(info parse.OpReturn, initialSync bool) error {
		if len(info.PushData) != 2 {
			if err := item.LogProcessError(&item.ProcessError{
				TxHash: info.TxHash,
				Error:  fmt.Sprintf("invalid set like, incorrect push data (%d)", len(info.PushData)),
			}); err != nil {
				return jerr.Get("error saving process error memo like incorrect push data", err)
			}
			return nil
		}
		if len(info.PushData[1]) != memo.TxHashLength {
			if err := item.LogProcessError(&item.ProcessError{
				TxHash: info.TxHash,
				Error:  fmt.Sprintf("error like tx hash not correct size: %d", len(info.PushData[1])),
			}); err != nil {
				return jerr.Get("error saving process error memo like post tx hash", err)
			}
			return nil
		}
		var postTxHash [32]byte
		copy(postTxHash[:], info.PushData[1])
		var memoLike = &dbMemo.AddrHeightLike{
			Addr:       info.Addr,
			Height:     info.Height,
			LikeTxHash: info.TxHash,
			PostTxHash: postTxHash,
		}
		var memoLiked = &dbMemo.PostHeightLike{
			PostTxHash: postTxHash,
			Height:     info.Height,
			LikeTxHash: info.TxHash,
			Addr:       info.Addr,
		}
		memoPost, err := dbMemo.GetPost(postTxHash)
		if err != nil {
			return jerr.Get("error getting memo post for like op return handler", err)
		}
		var objects = []db.Object{memoLike, memoLiked}
		if memoPost != nil && memoLike.Addr != memoPost.Addr {
			var tip int64
			for _, txOut := range info.Outputs {
				outputAddress, _ := wallet.GetAddrFromLockScript(txOut.PkScript)
				if outputAddress != nil && *outputAddress == memoPost.Addr {
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
		if !initialSync && info.Height != item.HeightMempool {
			memoLike.Height = item.HeightMempool
			if err := db.Remove([]db.Object{memoLike}); err != nil {
				return jerr.Get("error removing db memo like", err)
			}
		}
		return nil
	},
}
