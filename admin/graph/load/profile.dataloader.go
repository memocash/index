package load

import (
	"context"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/admin/graph/dataloader"
	"github.com/memocash/index/admin/graph/model"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/memo"
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
			addr, err := wallet.GetAddrFromString(addressString)
			if err != nil {
				errors[i] = jerr.Get("error getting address from string", err)
				continue
			}
			addrString := addr.String()
			var profile = &model.Profile{Address: addrString}
			addrMemoName, err := memo.GetAddrName(ctx, *addr)
			if err != nil && !client.IsEntryNotFoundError(err) {
				errors[i] = jerr.Get("error getting memo name", err)
				continue
			}
			if addrMemoName != nil {
				profile.Name = &model.SetName{
					TxHash:  chainhash.Hash(addrMemoName.TxHash).String(),
					Name:    addrMemoName.Name,
					Address: addrString,
				}
			}
			addrMemoProfile, err := memo.GetAddrProfile(ctx, *addr)
			if err != nil && !client.IsEntryNotFoundError(err) {
				errors[i] = jerr.Get("error getting addr memo profile", err)
				continue
			}
			if addrMemoProfile != nil {
				profile.Profile = &model.SetProfile{
					TxHash:  chainhash.Hash(addrMemoProfile.TxHash).String(),
					Text:    addrMemoProfile.Profile,
					Address: addrString,
				}
			}
			addrMemoProfilePic, err := memo.GetAddrProfilePic(ctx, *addr)
			if err != nil && !client.IsEntryNotFoundError(err) {
				errors[i] = jerr.Get("error getting addr memo profile pic", err)
				continue
			}
			if addrMemoProfilePic != nil {
				profile.Pic = &model.SetPic{
					TxHash:  chainhash.Hash(addrMemoProfilePic.TxHash).String(),
					Address: addrString,
					Pic:     addrMemoProfilePic.Pic,
				}
			}
			profiles[i] = profile
		}
		return profiles, errors
	},
}
