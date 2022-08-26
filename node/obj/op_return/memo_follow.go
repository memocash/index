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
	"github.com/memocash/index/ref/bitcoin/wallet"
)

var memoFollowHandler = &Handler{
	prefix: memo.PrefixFollow,
	handle: func(info parse.OpReturn) error {
		if len(info.PushData) != 2 {
			if err := item.LogProcessError(&item.ProcessError{
				TxHash: info.TxHash,
				Error:  fmt.Sprintf("invalid set follow, incorrect push data (%d)", len(info.PushData)),
			}); err != nil {
				return jerr.Get("error saving process error memo follow incorrect push data", err)
			}
			return nil
		}
		unfollow := bytes.Equal(info.PushData[0], memo.PrefixUnfollow)
		followAddress, err := wallet.GetAddressFromPkHashNew(info.PushData[1])
		if err != nil {
			if err := item.LogProcessError(&item.ProcessError{
				TxHash: info.TxHash,
				Error:  fmt.Sprintf("error getting address from follow pk hash: %s", err),
			}); err != nil {
				return jerr.Get("error saving process error memo follow address", err)
			}
			return nil
		}
		followLockHash := script.GetLockHashForAddress(followAddress)
		var lockMemoFollow = &dbMemo.LockHeightFollow{
			LockHash: info.LockHash,
			Height:   info.Height,
			TxHash:   info.TxHash,
			Follow:   followLockHash,
			Unfollow: unfollow,
		}
		var lockMemoFollowed = &dbMemo.LockHeightFollowed{
			FollowLockHash: followLockHash,
			Height:         info.Height,
			TxHash:         info.TxHash,
			LockHash:       info.LockHash,
			Unfollow:       unfollow,
		}
		var lockAddress = &item.LockAddress{
			LockHash: followLockHash,
			Address:  followAddress.GetEncoded(),
		}
		if err := db.Save([]db.Object{lockMemoFollow, lockMemoFollowed, lockAddress}); err != nil {
			return jerr.Get("error saving db lock memo follow object", err)
		}
		if info.Height != item.HeightMempool {
			lockMemoFollow.Height = item.HeightMempool
			if err := dbMemo.RemoveLockHeightFollow(lockMemoFollow); err != nil {
				return jerr.Get("error removing db lock memo follow", err)
			}
			lockMemoFollowed.Height = item.HeightMempool
			if err := dbMemo.RemoveLockHeightFollowed(lockMemoFollowed); err != nil {
				return jerr.Get("error removing db lock memo followed", err)
			}
		}
		return nil
	},
}

var memoUnfollowHandler = &Handler{
	prefix: memo.PrefixUnfollow,
	handle: memoFollowHandler.handle,
}
