package load

import (
	"context"
	"fmt"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/memo"
	"github.com/memocash/index/graph/model"
)

func GetProfile(ctx context.Context, address [25]byte) (*model.Profile, error) {
	var profile = &model.Profile{Address: address}
	addrMemoName, err := memo.GetAddrName(ctx, address)
	if err != nil && !client.IsEntryNotFoundError(err) {
		return nil, fmt.Errorf("error getting memo name for profile; %w", err)
	}
	if addrMemoName != nil {
		profile.Name = &model.SetName{
			TxHash:  addrMemoName.TxHash,
			Name:    addrMemoName.Name,
			Address: address,
		}
	}
	addrMemoProfile, err := memo.GetAddrProfile(ctx, address)
	if err != nil && !client.IsEntryNotFoundError(err) {
		return nil, fmt.Errorf("error getting memo profile text for profile; %w", err)
	}
	if addrMemoProfile != nil {
		profile.Profile = &model.SetProfile{
			TxHash:  addrMemoProfile.TxHash,
			Text:    addrMemoProfile.Profile,
			Address: address,
		}
	}
	addrMemoProfilePic, err := memo.GetAddrProfilePic(ctx, address)
	if err != nil && !client.IsEntryNotFoundError(err) {
		return nil, fmt.Errorf("error getting memo profile pic for profile; %w", err)
	}
	if addrMemoProfilePic != nil {
		profile.Pic = &model.SetPic{
			TxHash:  addrMemoProfilePic.TxHash,
			Address: address,
			Pic:     addrMemoProfilePic.Pic,
		}
	}
	return profile, nil
}
