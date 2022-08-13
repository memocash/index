package load

import (
	"context"
	"encoding/hex"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/admin/graph/dataloader"
	"github.com/memocash/index/admin/graph/model"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/tx/hs"
	"github.com/memocash/index/ref/bitcoin/tx/script"
	"github.com/memocash/index/ref/bitcoin/wallet"
	"time"
)

var ProfileLoaderConfig = dataloader.ProfileLoaderConfig{
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
			if err := db.Save([]db.Object{&item.LockAddress{
				LockHash: lockHash,
				Address:  address.GetEncoded(),
			}}); err != nil {
				errors[i] = jerr.Get("error saving lock address", err)
				continue
			}
			lockHashString := hex.EncodeToString(lockHash)
			var profile = &model.Profile{LockHash: lockHashString}
			lockMemoName, err := item.GetLockMemoName(ctx, lockHash)
			if err != nil && !client.IsEntryNotFoundError(err) {
				errors[i] = jerr.Get("error getting memo name", err)
				continue
			}
			if lockMemoName != nil {
				profile.Name = &model.SetName{
					TxHash:   hs.GetTxString(lockMemoName.TxHash),
					Name:     lockMemoName.Name,
					LockHash: lockHashString,
				}
			}
			lockMemoProfile, err := item.GetLockMemoProfile(ctx, lockHash)
			if err != nil && !client.IsEntryNotFoundError(err) {
				errors[i] = jerr.Get("error getting lock memo profile", err)
				continue
			}
			if lockMemoProfile != nil {
				profile.Profile = &model.SetProfile{
					TxHash:   hs.GetTxString(lockMemoProfile.TxHash),
					Text:     lockMemoProfile.Profile,
					LockHash: lockHashString,
				}
			}
			lockMemoProfilePic, err := item.GetLockMemoProfilePic(ctx, lockHash)
			if err != nil && !client.IsEntryNotFoundError(err) {
				errors[i] = jerr.Get("error getting lock memo profile pic", err)
				continue
			}
			if lockMemoProfilePic != nil {
				profile.Pic = &model.SetPic{
					TxHash:   hs.GetTxString(lockMemoProfilePic.TxHash),
					LockHash: lockHashString,
					Pic:      lockMemoProfilePic.Pic,
				}
			}
			profiles[i] = profile
		}
		return profiles, errors
	},
}
