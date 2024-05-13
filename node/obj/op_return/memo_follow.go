package op_return

import (
	"bytes"
	"context"
	"fmt"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/db/item/db"
	dbMemo "github.com/memocash/index/db/item/memo"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/parse"
	"github.com/memocash/index/ref/bitcoin/wallet"
)

var memoFollowHandler = &Handler{
	prefix: memo.PrefixFollow,
	handle: func(ctx context.Context, info parse.OpReturn) error {
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
				Error:  fmt.Sprintf("error getting address from follow pk hash; %s", err),
			}); err != nil {
				return jerr.Get("error saving process error memo follow address", err)
			}
			return nil
		}
		followAddr := followAddress.GetAddr()
		var addrMemoFollow = &dbMemo.AddrFollow{
			Addr:       info.Addr,
			Seen:       info.Seen,
			TxHash:     info.TxHash,
			FollowAddr: followAddr,
			Unfollow:   unfollow,
		}
		var addrMemoFollowed = &dbMemo.AddrFollowed{
			FollowAddr: followAddr,
			Seen:       info.Seen,
			TxHash:     info.TxHash,
			Addr:       info.Addr,
			Unfollow:   unfollow,
		}
		if err := db.Save([]db.Object{addrMemoFollow, addrMemoFollowed}); err != nil {
			return jerr.Get("error saving db lock memo follow object", err)
		}
		return nil
	},
}

var memoUnfollowHandler = &Handler{
	prefix: memo.PrefixUnfollow,
	handle: memoFollowHandler.handle,
}
