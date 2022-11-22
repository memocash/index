package save

import (
	"bytes"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/db/item/memo"
	"github.com/memocash/index/ref/bitcoin/tx/parse"
	"github.com/memocash/index/ref/bitcoin/tx/script"
)

func MemoPost(info parse.OpReturn, post string) error {
	var lockMemoPost = &memo.LockHeightPost{
		LockHash: info.LockHash,
		Height:   info.Height,
		TxHash:   info.TxHash,
	}
	var objects = []db.Object{lockMemoPost}
	existingMemoPost, err := memo.GetPost(info.TxHash)
	if err != nil {
		return jerr.Get("error getting existing memo post for post op return handler", err)
	}
	if existingMemoPost == nil {
		var memoPost = &memo.Post{
			TxHash:   info.TxHash,
			LockHash: info.LockHash,
			Post:     post,
		}
		objects = append(objects, memoPost)
		memoPostLikes, err := memo.GetPostHeightLikes([][]byte{info.TxHash})
		if err != nil {
			return jerr.Get("error getting memo likeds for post op return handler", err)
		}
		var likeTxHashes [][]byte
		for _, memoPostLike := range memoPostLikes {
			if !bytes.Equal(memoPostLike.LockHash, memoPost.LockHash) {
				likeTxHashes = append(likeTxHashes, memoPostLike.LikeTxHash)
			}
		}
		likeTxOuts, err := chain.GetTxOutputsByHashes(likeTxHashes)
		if err != nil {
			return jerr.Get("error getting like tx outputs for post op return handler", err)
		}
		var memoLikeTips = make(map[chainhash.Hash]int64)
		for _, likeTxOut := range likeTxOuts {
			lockHash := script.GetLockHash(likeTxOut.LockScript)
			if bytes.Equal(lockHash, memoPost.LockHash) {
				memoLikeTips[likeTxOut.TxHash] += likeTxOut.Value
			}
		}
		for likeTxHash, tip := range memoLikeTips {
			if tip > 0 {
				objects = append(objects, &memo.LikeTip{
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
		if err := memo.RemoveLockHeightPost(lockMemoPost); err != nil {
			return jerr.Get("error removing db memo post", err)
		}
	}
	return nil
}
