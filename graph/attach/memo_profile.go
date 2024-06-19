package attach

import (
	"context"
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/memo"
	"github.com/memocash/index/graph/model"
	"time"
)

type MemoProfile struct {
	base
	Profiles []*model.Profile
}

func ToMemoProfiles(ctx context.Context, fields []Field, profiles []*model.Profile) error {
	if len(profiles) == 0 {
		return nil
	}
	o := MemoProfile{
		base:     base{Ctx: ctx, Fields: fields},
		Profiles: profiles,
	}
	o.Wait.Add(8)
	go o.AttachLocks()
	go o.AttachPosts()
	go o.AttachFollowing()
	go o.AttachFollowers()
	go o.AttachRooms()
	go o.AttachNames()
	go o.AttachProfiles()
	go o.AttachPics()
	o.Wait.Wait()
	if len(o.Errors) > 0 {
		return fmt.Errorf("error attaching to memo profiles; %w", o.Errors[0])
	}
	return nil
}

func (a *MemoProfile) getAddresses() [][25]byte {
	a.Mutex.Lock()
	defer a.Mutex.Unlock()
	var addresses [][25]byte
	for i := range a.Profiles {
		addresses = append(addresses, a.Profiles[i].Address)
	}
	return addresses
}

func (a *MemoProfile) AttachLocks() {
	defer a.Wait.Done()
	var allLocks []*model.Lock
	if !a.HasField([]string{"lock"}) {
		return
	}
	a.Mutex.Lock()
	for _, profile := range a.Profiles {
		profile.Lock = &model.Lock{Address: profile.Address}
		allLocks = append(allLocks, profile.Lock)
	}
	a.Mutex.Unlock()
	if err := ToLocks(a.Ctx, GetPrefixFields(a.Fields, "lock."), allLocks); err != nil {
		a.AddError(fmt.Errorf("error attaching to locks for memo profiles; %w", err))
		return
	}
}

func (a *MemoProfile) AttachPosts() {
	defer a.Wait.Done()
	if !a.HasField([]string{"posts"}) {
		return
	}
	postsField := a.Fields.GetField("posts")
	startDate, _ := model.UnmarshalDate(postsField.Arguments["start"])
	newest, _ := graphql.UnmarshalBoolean(postsField.Arguments["newest"])
	var allProfilePosts []*model.Post
	for _, addr := range a.getAddresses() {
		addrPosts, err := memo.GetSingleAddrPosts(a.Ctx, addr, newest, time.Time(startDate))
		if err != nil && !client.IsEntryNotFoundError(err) {
			a.AddError(fmt.Errorf("error getting memo profile posts for profile attach; %w", err))
			return
		}
		a.Mutex.Lock()
		for _, profile := range a.Profiles {
			if profile.Address == addr {
				for _, addrPost := range addrPosts {
					post := &model.Post{
						TxHash: addrPost.TxHash,
					}
					profile.Posts = append(profile.Posts, post)
					allProfilePosts = append(allProfilePosts, post)
				}
			}
		}
		a.Mutex.Unlock()
	}
	if err := ToMemoPosts(a.Ctx, postsField.Fields, allProfilePosts); err != nil {
		a.AddError(fmt.Errorf("error attaching to posts for memo profiles; %w", err))
		return
	}
}

func (a *MemoProfile) AttachFollowing() {
	defer a.Wait.Done()
	if !a.HasField([]string{"following"}) {
		return
	}
	followingField := a.Fields.GetField("following")
	startDate, _ := model.UnmarshalDate(followingField.Arguments["start"])
	var allFollows []*model.Follow
	for _, addr := range a.getAddresses() {
		addrMemoFollows, err := memo.GetAddrFollowsSingle(a.Ctx, addr, time.Time(startDate))
		if err != nil {
			a.AddError(fmt.Errorf("error getting address memo follows for address; %w", err))
			return
		}
		a.Mutex.Lock()
		for _, profile := range a.Profiles {
			if profile.Address == addr {
				for _, addrMemoFollow := range addrMemoFollows {
					follow := &model.Follow{
						Address:       addrMemoFollow.Addr,
						TxHash:        addrMemoFollow.TxHash,
						Unfollow:      addrMemoFollow.Unfollow,
						FollowAddress: addrMemoFollow.FollowAddr,
					}
					profile.Following = append(profile.Following, follow)
					allFollows = append(allFollows, follow)
				}
			}
		}
		a.Mutex.Unlock()
	}
	if err := ToMemoFollows(a.Ctx, followingField.Fields, allFollows); err != nil {
		a.AddError(fmt.Errorf("error attaching to following for memo profiles; %w", err))
		return
	}
}

func (a *MemoProfile) AttachFollowers() {
	defer a.Wait.Done()
	if !a.HasField([]string{"followers"}) {
		return
	}
	followersField := a.Fields.GetField("followers")
	startDate, _ := model.UnmarshalDate(followersField.Arguments["start"])
	var allFollows []*model.Follow
	for _, addr := range a.getAddresses() {
		addrMemoFollows, err := memo.GetAddrFollowedsSingle(a.Ctx, addr, time.Time(startDate))
		if err != nil {
			a.AddError(fmt.Errorf("error getting address memo followeds for address; %w", err))
			return
		}
		a.Mutex.Lock()
		for _, profile := range a.Profiles {
			if profile.Address == addr {
				for _, addrMemoFollow := range addrMemoFollows {
					follow := &model.Follow{
						Address:       addrMemoFollow.Addr,
						TxHash:        addrMemoFollow.TxHash,
						Unfollow:      addrMemoFollow.Unfollow,
						FollowAddress: addrMemoFollow.FollowAddr,
					}
					profile.Followers = append(profile.Followers, follow)
					allFollows = append(allFollows, follow)
				}
			}
		}
		a.Mutex.Unlock()
	}
	if err := ToMemoFollows(a.Ctx, followersField.Fields, allFollows); err != nil {
		a.AddError(fmt.Errorf("error attaching to followers for memo profiles; %w", err))
		return
	}
}

func (a *MemoProfile) AttachRooms() {
	defer a.Wait.Done()
	if !a.HasField([]string{"rooms"}) {
		return
	}
	lockRoomFollows, err := memo.GetAddrRoomFollows(a.Ctx, a.getAddresses())
	if err != nil {
		a.AddError(fmt.Errorf("error getting addr room follows for profile attach; %w", err))
		return
	}
	var allRoomFollows []*model.RoomFollow
	a.Mutex.Lock()
	for _, lockRoomFollow := range lockRoomFollows {
		if lockRoomFollow.Unfollow {
			continue
		}
		for _, profile := range a.Profiles {
			if profile.Address == lockRoomFollow.Addr {
				roomFollow := &model.RoomFollow{
					Name:     lockRoomFollow.Room,
					Address:  lockRoomFollow.Addr,
					TxHash:   lockRoomFollow.TxHash,
					Unfollow: lockRoomFollow.Unfollow,
				}
				profile.Rooms = append(profile.Rooms, roomFollow)
				allRoomFollows = append(allRoomFollows, roomFollow)
			}
		}
	}
	a.Mutex.Unlock()
	if err := ToMemoRoomFollows(a.Ctx, GetPrefixFields(a.Fields, "rooms"), allRoomFollows); err != nil {
		a.AddError(fmt.Errorf("error attaching to rooms for memo profiles; %w", err))
		return
	}
}

func (a *MemoProfile) AttachNames() {
	defer a.Wait.Done()
	if !a.HasField([]string{"name"}) {
		return
	}
	var allSetNames []*model.SetName
	for _, address := range a.getAddresses() {
		allSetNames = append(allSetNames, &model.SetName{Address: address})
	}
	if err := ToMemoSetNames(a.Ctx, GetPrefixFields(a.Fields, "name"), allSetNames); err != nil {
		a.AddError(fmt.Errorf("error attaching to names for memo profiles; %w", err))
		return
	}
	a.Mutex.Lock()
	defer a.Mutex.Unlock()
	for _, setName := range allSetNames {
		if !SetNameExists(setName) {
			continue
		}
		for _, profile := range a.Profiles {
			if profile.Address == setName.Address {
				profile.Name = setName
			}
		}
	}
}

func (a *MemoProfile) AttachProfiles() {
	defer a.Wait.Done()
	if !a.HasField([]string{"profile"}) {
		return
	}
	var allSetProfiles []*model.SetProfile
	for _, address := range a.getAddresses() {
		allSetProfiles = append(allSetProfiles, &model.SetProfile{Address: address})
	}
	if err := ToMemoSetProfiles(a.Ctx, GetPrefixFields(a.Fields, "profile"), allSetProfiles); err != nil {
		a.AddError(fmt.Errorf("error attaching to profiles for memo profiles; %w", err))
		return
	}
	a.Mutex.Lock()
	defer a.Mutex.Unlock()
	for _, setProfile := range allSetProfiles {
		if !SetProfileExists(setProfile) {
			continue
		}
		for _, profile := range a.Profiles {
			if profile.Address == setProfile.Address {
				profile.Profile = setProfile
			}
		}
	}
}

func (a *MemoProfile) AttachPics() {
	defer a.Wait.Done()
	if !a.HasField([]string{"pic"}) {
		return
	}
	var allSetPics []*model.SetPic
	for _, address := range a.getAddresses() {
		allSetPics = append(allSetPics, &model.SetPic{Address: address})
	}
	if err := ToMemoSetPics(a.Ctx, GetPrefixFields(a.Fields, "pic"), allSetPics); err != nil {
		a.AddError(fmt.Errorf("error attaching to pics for memo profiles; %w", err))
		return
	}
	a.Mutex.Lock()
	defer a.Mutex.Unlock()
	for _, setPic := range allSetPics {
		if !SetPicExists(setPic) {
			continue
		}
		for _, profile := range a.Profiles {
			if profile.Address == setPic.Address {
				profile.Pic = setPic
			}
		}
	}
}

func (a *MemoProfile) AttachLinks() {
	defer a.Wait.Done()
	if !a.HasField([]string{"links"}) {
		return
	}
	var allAddresses [][25]byte
	var allLinkRequests []*memo.LinkRequest
	var checkAddresses = a.getAddresses()
	for len(checkAddresses) > 0 {
		addrLinkRequests, err := memo.GetAddrLinkRequests(a.Ctx, checkAddresses)
		if err != nil {
			a.AddError(fmt.Errorf("error getting addr link requests for profile attach; %w", err))
			return
		}
		addrLinkRequesteds, err := memo.GetAddrLinkRequesteds(a.Ctx, checkAddresses)
		if err != nil {
			a.AddError(fmt.Errorf("error getting addr link requesteds for profile attach; %w", err))
			return
		}
		var linkRequestTxHashes [][32]byte
		for _, addrLinkRequest := range addrLinkRequests {
			linkRequestTxHashes = append(linkRequestTxHashes, addrLinkRequest.TxHash)
		}
		for _, addrLinkRequested := range addrLinkRequesteds {
			linkRequestTxHashes = append(linkRequestTxHashes, addrLinkRequested.TxHash)
		}
		linkRequests, err := memo.GetLinkRequests(a.Ctx, linkRequestTxHashes)
		if err != nil {
			a.AddError(fmt.Errorf("error getting link requests for profile attach; %w", err))
			return
		}
		for _, linkRequest := range linkRequests {
			var found bool
			for _, allLinkRequest := range allLinkRequests {
				if linkRequest.TxHash == allLinkRequest.TxHash {
					found = true
					break
				}
			}
			if !found {
				allLinkRequests = append(allLinkRequests, linkRequest)
			}
		}
		allAddresses = checkAddresses
		checkAddresses = nil
		for _, linkRequest := range linkRequests {
			var childAddrFound, parentAddrFound bool
			for _, addr := range allAddresses {
				if linkRequest.ChildAddr == addr {
					childAddrFound = true
				} else if linkRequest.ParentAddr == addr {
					parentAddrFound = true
				}
			}
			if !childAddrFound {
				checkAddresses = append(checkAddresses, linkRequest.ChildAddr)
			} else if !parentAddrFound {
				checkAddresses = append(checkAddresses, linkRequest.ParentAddr)
			}
		}
	}
	a.Mutex.Lock()
	defer a.Mutex.Unlock()
	for _, profile := range a.Profiles {
		var foundLinkRequests []*memo.LinkRequest
		for _, linkRequest := range allLinkRequests {
			if linkRequest.ChildAddr == profile.Address || linkRequest.ParentAddr == profile.Address {
				foundLinkRequests = append(foundLinkRequests, linkRequest)
			}
		}
		for more := true; more; {
			more = false
			for _, linkRequest := range allLinkRequests {
				var shouldBeIncluded, alreadyIncluded bool
				for _, foundLinkRequest := range foundLinkRequests {
					if foundLinkRequest.TxHash == linkRequest.TxHash {
						alreadyIncluded = true
						break
					}
					if linkRequest.ChildAddr == foundLinkRequest.ParentAddr ||
						linkRequest.ParentAddr == foundLinkRequest.ChildAddr {
						shouldBeIncluded = true
					}
				}
				if alreadyIncluded {
					continue
				} else if shouldBeIncluded {
					foundLinkRequests = append(foundLinkRequests, linkRequest)
					more = true
				}
			}
		}
		for _, linkRequest := range foundLinkRequests {
			profile.Links = append(profile.Links, &model.Link{
				Request: &model.LinkRequest{
					Child:   &model.Lock{Address: linkRequest.ChildAddr},
					Parent:  &model.Lock{Address: linkRequest.ParentAddr},
					Message: linkRequest.Message,
					Tx:      &model.Tx{Hash: linkRequest.TxHash},
				},
			})
		}
	}
}
