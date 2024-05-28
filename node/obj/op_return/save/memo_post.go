package save

import (
	"context"
	"fmt"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/db/item/memo"
	"github.com/memocash/index/ref/bitcoin/tx/parse"
	"github.com/memocash/index/ref/bitcoin/wallet"
)

func MemoPost(ctx context.Context, info parse.OpReturn, post string) error {
	var lockMemoPost = &memo.AddrPost{
		Addr:   info.Addr,
		Seen:   info.Seen,
		TxHash: info.TxHash,
	}
	var memoSeenPost = &memo.SeenPost{
		Seen:       info.Seen,
		PostTxHash: info.TxHash,
	}
	var objects = []db.Object{lockMemoPost, memoSeenPost}
	existingMemoPost, err := memo.GetPost(ctx, info.TxHash)
	if err != nil {
		return fmt.Errorf("error getting existing memo post for post op return handler; %w", err)
	}
	if existingMemoPost == nil {
		var memoPost = &memo.Post{
			TxHash: info.TxHash,
			Addr:   info.Addr,
			Post:   post,
		}
		objects = append(objects, memoPost)
		memoPostLikes, err := memo.GetPostLikes(ctx, [][32]byte{info.TxHash})
		if err != nil {
			return fmt.Errorf("error getting memo likeds for post op return handler; %w", err)
		}
		var likeTxHashes [][32]byte
		for _, memoPostLike := range memoPostLikes {
			if memoPostLike.Addr != memoPost.Addr {
				likeTxHashes = append(likeTxHashes, memoPostLike.LikeTxHash)
			}
		}
		likeTxOuts, err := chain.GetTxOutputsByHashes(ctx, likeTxHashes)
		if err != nil {
			return fmt.Errorf("error getting like tx outputs for post op return handler; %w", err)
		}
		var memoLikeTips = make(map[chainhash.Hash]int64)
		for _, likeTxOut := range likeTxOuts {
			addr, _ := wallet.GetAddrFromLockScript(likeTxOut.LockScript)
			if addr != nil && *addr == memoPost.Addr {
				memoLikeTips[likeTxOut.TxHash] += likeTxOut.Value
			}
		}
		for likeTxHash, tip := range memoLikeTips {
			if tip > 0 {
				objects = append(objects, &memo.LikeTip{
					LikeTxHash: likeTxHash,
					Tip:        tip,
				})
			}
		}
	}
	if err := db.Save(objects); err != nil {
		return fmt.Errorf("error saving db memo post object; %w", err)
	}
	return nil
}
