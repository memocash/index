package resolver

import (
	"context"
	"encoding/hex"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/admin/graph/dataloader"
	"github.com/memocash/index/admin/graph/model"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/ref/bitcoin/tx/hs"
	"github.com/memocash/index/ref/bitcoin/tx/script"
	"github.com/memocash/index/ref/bitcoin/wallet"
	"time"
)

var profileLoaderConfig = dataloader.ProfileLoaderConfig{
	Wait:     2 * time.Millisecond,
	MaxBatch: 100,
	Fetch: func(addressStrings []string) ([]*model.Profile, []error) {
		var profiles = make([]*model.Profile, len(addressStrings))
		var errors = make([]error, len(addressStrings))
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		for i, addressString := range addressStrings {
			address := wallet.GetAddressFromString(addressString)
			lockHash := script.GetLockHashForAddress(address)
			if err := item.Save([]item.Object{&item.LockAddress{
				LockHash: lockHash,
				Address:  address.GetEncoded(),
			}}); err != nil {
				errors[i] = jerr.Get("error saving lock address", err)
				continue
			}
			lockHashString := hex.EncodeToString(lockHash)
			var profile = &model.Profile{LockHash: lockHashString}
			memoName, err := item.GetMemoName(ctx, lockHash)
			if err != nil && !client.IsEntryNotFoundError(err) {
				errors[i] = jerr.Get("error getting memo name", err)
				continue
			}
			if memoName != nil {
				profile.Name = &model.SetName{
					TxHash:   hs.GetTxString(memoName.TxHash),
					Name:     memoName.Name,
					LockHash: lockHashString,
				}
			}
			memoProfile, err := item.GetMemoProfile(ctx, lockHash)
			if err != nil && !client.IsEntryNotFoundError(err) {
				errors[i] = jerr.Get("error getting memo profile", err)
				continue
			}
			if memoProfile != nil {
				profile.Profile = &model.SetProfile{
					TxHash:   hs.GetTxString(memoProfile.TxHash),
					Text:     memoProfile.Profile,
					LockHash: lockHashString,
				}
			}
			memoProfilePic, err := item.GetMemoProfilePic(ctx, lockHash)
			if err != nil && !client.IsEntryNotFoundError(err) {
				errors[i] = jerr.Get("error getting memo profile pic", err)
				continue
			}
			if memoProfilePic != nil {
				profile.Pic = &model.SetPic{
					TxHash:   hs.GetTxString(memoProfilePic.TxHash),
					LockHash: lockHashString,
					Pic:      memoProfilePic.Pic,
				}
			}
			profiles[i] = profile
		}
		return profiles, errors
	},
}
