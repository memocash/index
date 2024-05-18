package load

import (
	"context"
	"fmt"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/memo"
	"github.com/memocash/index/graph/model"
	"github.com/memocash/index/ref/bitcoin/wallet"
)

func GetProfile(ctx context.Context, addressString string) (*model.Profile, error) {
	address, err := wallet.GetAddrFromString(addressString)
	if err != nil {
		return nil, fmt.Errorf("error getting address from profile dataloader: %s; %w", addressString, err)
	}
	var profile = &model.Profile{Address: addressString}
	addrMemoName, err := memo.GetAddrName(ctx, *address)
	if err != nil && !client.IsEntryNotFoundError(err) {
		return nil, fmt.Errorf("error getting memo name for profile; %w", err)
	}
	if addrMemoName != nil {
		profile.Name = &model.SetName{
			TxHash:  addrMemoName.TxHash,
			Name:    addrMemoName.Name,
			Address: addressString,
		}
	}
	addrMemoProfile, err := memo.GetAddrProfile(ctx, *address)
	if err != nil && !client.IsEntryNotFoundError(err) {
		return nil, fmt.Errorf("error getting memo profile text for profile; %w", err)
	}
	if addrMemoProfile != nil {
		profile.Profile = &model.SetProfile{
			TxHash:  addrMemoProfile.TxHash,
			Text:    addrMemoProfile.Profile,
			Address: addressString,
		}
	}
	addrMemoProfilePic, err := memo.GetAddrProfilePic(ctx, *address)
	if err != nil && !client.IsEntryNotFoundError(err) {
		return nil, fmt.Errorf("error getting memo profile pic for profile; %w", err)
	}
	if addrMemoProfilePic != nil {
		profile.Pic = &model.SetPic{
			TxHash:  addrMemoProfilePic.TxHash,
			Address: addressString,
			Pic:     addrMemoProfilePic.Pic,
		}
	}
	return profile, nil
}
