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
	a.Mutex.Lock()
	for _, profile := range a.Profiles {
		profile.Name = &model.SetName{Address: profile.Address}
		allSetNames = append(allSetNames, profile.Name)
	}
	a.Mutex.Unlock()
	if err := ToMemoSetNames(a.Ctx, GetPrefixFields(a.Fields, "name"), allSetNames); err != nil {
		a.AddError(fmt.Errorf("error attaching to names for memo profiles; %w", err))
		return
	}
}

func (a *MemoProfile) AttachProfiles() {
	defer a.Wait.Done()
	if !a.HasField([]string{"profile"}) {
		return
	}
	var allSetProfiles []*model.SetProfile
	a.Mutex.Lock()
	for _, profile := range a.Profiles {
		profile.Profile = &model.SetProfile{Address: profile.Address}
		allSetProfiles = append(allSetProfiles, profile.Profile)
	}
	a.Mutex.Unlock()
	if err := ToMemoSetProfiles(a.Ctx, GetPrefixFields(a.Fields, "profile"), allSetProfiles); err != nil {
		a.AddError(fmt.Errorf("error attaching to profiles for memo profiles; %w", err))
		return
	}
}

func (a *MemoProfile) AttachPics() {
	defer a.Wait.Done()
	if !a.HasField([]string{"pic"}) {
		return
	}
	var allSetPics []*model.SetPic
	a.Mutex.Lock()
	for _, profile := range a.Profiles {
		profile.Pic = &model.SetPic{Address: profile.Address}
		allSetPics = append(allSetPics, profile.Pic)
	}
	a.Mutex.Unlock()
	if err := ToMemoSetPics(a.Ctx, GetPrefixFields(a.Fields, "pic"), allSetPics); err != nil {
		a.AddError(fmt.Errorf("error attaching to pics for memo profiles; %w", err))
		return
	}
}
