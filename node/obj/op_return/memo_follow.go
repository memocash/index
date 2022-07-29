package op_return

import (
	"bytes"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/script"
	"github.com/memocash/index/ref/bitcoin/wallet"
)

var memoFollowHandler = &Handler{
	prefix: memo.PrefixFollow,
	handle: func(info Info) error {
		if len(info.PushData) != 2 {
			return jerr.Newf("invalid set follow, incorrect push data (%d)", len(info.PushData))
		}
		unfollow := bytes.Equal(info.PushData[0], memo.PrefixUnfollow)
		followAddress := wallet.GetAddressFromPkHash(info.PushData[1])
		followLockHash := script.GetLockHashForAddress(followAddress)
		var memoFollow = &item.MemoFollow{
			LockHash: info.LockHash,
			Height:   info.Height,
			TxHash:   info.TxHash,
			Follow:   followLockHash,
			Unfollow: unfollow,
		}
		var memoFollowed = &item.MemoFollowed{
			FollowLockHash: followLockHash,
			Height:         info.Height,
			TxHash:         info.TxHash,
			LockHash:       info.LockHash,
			Unfollow:       unfollow,
		}
		var lockAddress = &item.LockAddress{
			LockHash: info.LockHash,
			Address:  followAddress.GetEncoded(),
		}
		if err := item.Save([]item.Object{memoFollow, memoFollowed, lockAddress}); err != nil {
			return jerr.Get("error saving db memo follow object", err)
		}
		if info.Height != item.HeightMempool {
			memoFollow.Height = item.HeightMempool
			if err := item.RemoveMemoFollow(memoFollow); err != nil {
				return jerr.Get("error removing db memo follow", err)
			}
			memoFollowed.Height = item.HeightMempool
			if err := item.RemoveMemoFollowed(memoFollowed); err != nil {
				return jerr.Get("error removing db memo followed", err)
			}
		}
		return nil
	},
}

var memoUnfollowHandler = &Handler{
	prefix: memo.PrefixUnfollow,
	handle: memoFollowHandler.handle,
}
