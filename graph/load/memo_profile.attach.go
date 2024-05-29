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
	o.Wait.Add(3)
	go o.AttachLocks()
	go o.AttachPosts()
	go o.AttachFollowing()
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
	postsField := a.Fields.GetField("following")
	startDate, _ := model.UnmarshalDate(postsField.Arguments["start"])
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
	if err := AttachToMemoFollows(a.Ctx, postsField.Fields, allFollows); err != nil {
		a.AddError(fmt.Errorf("error attaching to following for memo profiles; %w", err))
		return
	}
}
