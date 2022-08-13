package save

import (
	"bytes"
	"encoding/hex"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/tx/parse"
)

func MemoPost(info parse.OpReturn, post string) error {
	var lockMemoPost = &item.LockMemoPost{
		LockHash: info.LockHash,
		Height:   info.Height,
		TxHash:   info.TxHash,
	}
	var objects = []db.Object{lockMemoPost}
	existingMemoPost, err := item.GetMemoPost(info.TxHash)
	if err != nil {
		return jerr.Get("error getting existing memo post for post op return handler", err)
	}
	if existingMemoPost == nil {
		var memoPost = &item.MemoPost{
			TxHash:   info.TxHash,
			LockHash: info.LockHash,
			Post:     post,
		}
		objects = append(objects, memoPost)
		memoLikeds, err := item.GetMemoLikeds([][]byte{info.TxHash})
		if err != nil {
			return jerr.Get("error getting memo likeds for post op return handler", err)
		}
		var likeTxHashes [][]byte
		for _, memoLiked := range memoLikeds {
			if !bytes.Equal(memoLiked.LockHash, memoPost.LockHash) {
				likeTxHashes = append(likeTxHashes, memoLiked.LikeTxHash)
			}
		}
		likeTxOuts, err := item.GetTxOutputsByHashes(likeTxHashes)
		if err != nil {
			return jerr.Get("error getting like tx outputs for post op return handler", err)
		}
		var memoLikeTips = make(map[string]int64)
		for _, likeTxOut := range likeTxOuts {
			if bytes.Equal(likeTxOut.LockHash, memoPost.LockHash) {
				memoLikeTips[hex.EncodeToString(likeTxOut.TxHash)] += likeTxOut.Value
			}
		}
		for likeTxHashString, tip := range memoLikeTips {
			likeTxHash, err := chainhash.NewHashFromStr(likeTxHashString)
			if err != nil {
				return jerr.Get("error parsing like tip tx hash for memo post op return handler", err)
			}
			if tip > 0 {
				objects = append(objects, &item.MemoLikeTip{
					LikeTxHash: likeTxHash.CloneBytes(),
					Tip:        tip,
				})
			}
		}
	}
	if err := db.Save(objects); err != nil {
		return jerr.Get("error saving db memo post object", err)
	}
	if info.Height != item.HeightMempool {
		lockMemoPost.Height = item.HeightMempool
		if err := item.RemoveLockMemoPost(lockMemoPost); err != nil {
			return jerr.Get("error removing db memo post", err)
		}
	}
	return nil
}
