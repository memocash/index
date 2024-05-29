package load

import (
	"context"
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/memo"
	"github.com/memocash/index/graph/model"
	"time"
)

type MemoProfileAttach struct {
	baseA
	Profiles []*model.Profile
}

func AttachToMemoProfiles(ctx context.Context, fields []Field, profiles []*model.Profile) error {
	if len(profiles) == 0 {
		return nil
	}
	o := MemoProfileAttach{
		baseA:    baseA{Ctx: ctx, Fields: fields},
		Profiles: profiles,
	}
	o.Wait.Add(5)
	go o.AttachLocks()
	go o.AttachPosts()
	go o.AttachFollowing()
	go o.AttachFollowers()
	go o.AttachRooms()
	o.Wait.Wait()
	if len(o.Errors) > 0 {
		return fmt.Errorf("error attaching to memo profiles; %w", o.Errors[0])
	}
	return nil
}

func (a *MemoProfileAttach) getAddresses() [][25]byte {
	a.Mutex.Lock()
	defer a.Mutex.Unlock()
	var addresses [][25]byte
	for i := range a.Profiles {
		addresses = append(addresses, a.Profiles[i].Address)
	}
	return addresses
}

func (a *MemoProfileAttach) AttachLocks() {
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
	if err := AttachToLocks(a.Ctx, GetPrefixFields(a.Fields, "lock."), allLocks); err != nil {
		a.AddError(fmt.Errorf("error attaching to locks for memo profiles; %w", err))
		return
	}
}

func (a *MemoProfileAttach) AttachPosts() {
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
	if err := AttachToMemoPosts(a.Ctx, postsField.Fields, allProfilePosts); err != nil {
		a.AddError(fmt.Errorf("error attaching to posts for memo profiles; %w", err))
		return
	}
}

func (a *MemoProfileAttach) AttachFollowing() {
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
	if err := AttachToMemoFollows(a.Ctx, followingField.Fields, allFollows); err != nil {
		a.AddError(fmt.Errorf("error attaching to following for memo profiles; %w", err))
		return
	}
}

func (a *MemoProfileAttach) AttachFollowers() {
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
	if err := AttachToMemoFollows(a.Ctx, followersField.Fields, allFollows); err != nil {
		a.AddError(fmt.Errorf("error attaching to followers for memo profiles; %w", err))
		return
	}
}

func (a *MemoProfileAttach) AttachRooms() {
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
	if err := AttachToMemoRoomFollows(a.Ctx, GetPrefixFields(a.Fields, "rooms"), allRoomFollows); err != nil {
		a.AddError(fmt.Errorf("error attaching to rooms for memo profiles; %w", err))
		return
	}
}

func (a *MemoProfileAttach) AttachNames() {
	defer a.Wait.Done()
	if !a.HasField([]string{"name"}) {
		return
	}
	addrProfileNames, err := memo.GetAddrNames(a.Ctx, a.getAddresses())
	if err != nil {
		a.AddError(fmt.Errorf("error getting addr profile names for profile attach; %w", err))
		return
	}
	var allModelSetNames []*model.SetName
	a.Mutex.Lock()
	for _, addrProfileName := range addrProfileNames {
		for _, profile := range a.Profiles {
			if profile.Address == addrProfileName.Addr {
				profile.Name = &model.SetName{
					Address: addrProfileName.Addr,
					TxHash:  addrProfileName.TxHash,
					Name:    addrProfileName.Name,
				}
				allModelSetNames = append(allModelSetNames, profile.Name)
			}
		}
	}
	a.Mutex.Unlock()
	/*if err := AttachToMemoSetNames(a.Ctx, GetPrefixFields(a.Fields, "name"), allModelSetNames); err != nil {
		a.AddError(fmt.Errorf("error attaching to names for memo profiles; %w", err))
		return
	}*/
}

func (a *MemoProfileAttach) AttachProfiles() {
	defer a.Wait.Done()
	if !a.HasField([]string{"profile"}) {
		return
	}
	addrProfiles, err := memo.GetAddrProfiles(a.Ctx, a.getAddresses())
	if err != nil {
		a.AddError(fmt.Errorf("error getting addr profiles for profile attach; %w", err))
		return
	}
	var allModelSetProfiles []*model.SetProfile
	a.Mutex.Lock()
	for _, addrProfile := range addrProfiles {
		for _, profile := range a.Profiles {
			if profile.Address == addrProfile.Addr {
				profile.Profile = &model.SetProfile{
					Address: addrProfile.Addr,
					TxHash:  addrProfile.TxHash,
					Text:    addrProfile.Profile,
				}
				allModelSetProfiles = append(allModelSetProfiles, profile.Profile)
			}
		}
	}
	a.Mutex.Unlock()
	/*if err := AttachToMemoSetProfiles(a.Ctx, GetPrefixFields(a.Fields, "profile"), allModelSetProfiles); err != nil {
		a.AddError(fmt.Errorf("error attaching to profiles for memo profiles; %w", err))
		return
	}*/
}

func (a *MemoProfileAttach) AttachPics() {
	defer a.Wait.Done()
	if !a.HasField([]string{"pic"}) {
		return
	}
	addrProfilePics, err := memo.GetAddrProfilePics(a.Ctx, a.getAddresses())
	if err != nil {
		a.AddError(fmt.Errorf("error getting addr profile pics for profile attach; %w", err))
		return
	}
	var allModelSetPics []*model.SetPic
	a.Mutex.Lock()
	for _, addrProfilePic := range addrProfilePics {
		for _, profile := range a.Profiles {
			if profile.Address == addrProfilePic.Addr {
				profile.Pic = &model.SetPic{
					Address: addrProfilePic.Addr,
					TxHash:  addrProfilePic.TxHash,
					Pic:     addrProfilePic.Pic,
				}
				allModelSetPics = append(allModelSetPics, profile.Pic)
			}
		}
	}
	a.Mutex.Unlock()
	/*if err := AttachToMemoSetPics(a.Ctx, GetPrefixFields(a.Fields, "pic"), allModelSetPics); err != nil {
		a.AddError(fmt.Errorf("error attaching to pics for memo profiles; %w", err))
		return
	}*/
}
