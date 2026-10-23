package attach

import (
	"bytes"
	"context"
	"fmt"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/memo"
	"github.com/memocash/index/graph/model"
	"sort"
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
	o.Wait.Add(9)
	go o.AttachLocks()
	go o.AttachPosts()
	go o.AttachFollowing()
	go o.AttachFollowers()
	go o.AttachRooms()
	go o.AttachNames()
	go o.AttachProfiles()
	go o.AttachPics()
	go o.AttachLinks()
	o.Wait.Wait()
	if len(o.Errors) > 0 {
		return fmt.Errorf("error attaching to memo profiles; %w", o.Errors[0])
	}
	return nil
}

// getAddresses returns the distinct memo keys of the profiles, in first
// order. Two profiles can share a key (a P2SH32 address and its alias, or
// the same address twice); each key is read once and its rows handed to
// every profile under it through getProfileIndexMap, so no profile gets a
// row more than once.
func (a *MemoProfile) getAddresses() [][25]byte {
	a.Mutex.Lock()
	defer a.Mutex.Unlock()
	var addresses [][25]byte
	var seen = make(map[[25]byte]struct{}, len(a.Profiles))
	for i := range a.Profiles {
		address := a.Profiles[i].Address.Addr()
		if _, ok := seen[address]; ok {
			continue
		}
		seen[address] = struct{}{}
		addresses = append(addresses, address)
	}
	return addresses
}

func (a *MemoProfile) getProfileIndexMap() map[[25]byte][]int {
	a.Mutex.Lock()
	defer a.Mutex.Unlock()
	m := make(map[[25]byte][]int, len(a.Profiles))
	for i := range a.Profiles {
		m[a.Profiles[i].Address.Addr()] = append(m[a.Profiles[i].Address.Addr()], i)
	}
	return m
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
	limit, err := unmarshalPageLimit(postsField.Arguments, "profile posts", memo.MaxPageSize)
	if err != nil {
		a.AddError(err)
		return
	}
	newest := unmarshalBooleanDefault(postsField.Arguments, "newest", true)
	profileIndexMap := a.getProfileIndexMap()
	var allProfilePosts []*model.Post
	for _, addr := range a.getAddresses() {
		addrPosts, err := memo.GetSingleAddrPosts(a.Ctx, addr, newest, time.Time(startDate), limit)
		if err != nil && !client.IsEntryNotFoundError(err) {
			a.AddError(fmt.Errorf("error getting memo profile posts for profile attach; %w", err))
			return
		}
		a.Mutex.Lock()
		for _, i := range profileIndexMap[addr] {
			for _, addrPost := range addrPosts {
				post := &model.Post{
					TxHash: addrPost.TxHash,
				}
				a.Profiles[i].Posts = append(a.Profiles[i].Posts, post)
				allProfilePosts = append(allProfilePosts, post)
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
	limit, err := unmarshalPageLimit(followingField.Arguments, "profile following", memo.MaxPageSize)
	if err != nil {
		a.AddError(err)
		return
	}
	profileIndexMap := a.getProfileIndexMap()
	var allFollows []*model.Follow
	for _, addr := range a.getAddresses() {
		addrMemoFollows, err := memo.GetAddrFollowsSingle(a.Ctx, addr, time.Time(startDate), limit)
		if err != nil {
			a.AddError(fmt.Errorf("error getting address memo follows for address; %w", err))
			return
		}
		a.Mutex.Lock()
		for _, i := range profileIndexMap[addr] {
			for _, addrMemoFollow := range addrMemoFollows {
				follow := &model.Follow{
					Address:       model.NewAddress(addrMemoFollow.Addr),
					TxHash:        addrMemoFollow.TxHash,
					Unfollow:      addrMemoFollow.Unfollow,
					FollowAddress: model.NewAddress(addrMemoFollow.FollowAddr),
				}
				a.Profiles[i].Following = append(a.Profiles[i].Following, follow)
				allFollows = append(allFollows, follow)
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
	limit, err := unmarshalPageLimit(followersField.Arguments, "profile followers", memo.MaxPageSize)
	if err != nil {
		a.AddError(err)
		return
	}
	profileIndexMap := a.getProfileIndexMap()
	var allFollows []*model.Follow
	for _, addr := range a.getAddresses() {
		addrMemoFollows, err := memo.GetAddrFollowedsSingle(a.Ctx, addr, time.Time(startDate), limit)
		if err != nil {
			a.AddError(fmt.Errorf("error getting address memo followeds for address; %w", err))
			return
		}
		a.Mutex.Lock()
		for _, i := range profileIndexMap[addr] {
			for _, addrMemoFollow := range addrMemoFollows {
				follow := &model.Follow{
					Address:       model.NewAddress(addrMemoFollow.Addr),
					TxHash:        addrMemoFollow.TxHash,
					Unfollow:      addrMemoFollow.Unfollow,
					FollowAddress: model.NewAddress(addrMemoFollow.FollowAddr),
				}
				a.Profiles[i].Followers = append(a.Profiles[i].Followers, follow)
				allFollows = append(allFollows, follow)
			}
		}
		a.Mutex.Unlock()
	}
	if err := ToMemoFollows(a.Ctx, followersField.Fields, allFollows); err != nil {
		a.AddError(fmt.Errorf("error attaching to followers for memo profiles; %w", err))
		return
	}
}

func (a *MemoProfile) AttachLinks() {
	defer a.Wait.Done()
	if !a.HasField([]string{"links"}) {
		return
	}
	linkRequestsField := a.Fields.GetField("links")
	startDate, _ := model.UnmarshalDate(linkRequestsField.Arguments["start"])
	limit, err := unmarshalPageLimit(linkRequestsField.Arguments, "profile links", memo.MaxPageSize)
	if err != nil {
		a.AddError(err)
		return
	}
	limit = memo.NormalizePageLimit(limit)
	profileIndexMap := a.getProfileIndexMap()
	var allLinkRequests []*model.LinkRequest
	for _, addr := range a.getAddresses() {
		addrLinkRequests, err := memo.GetAddrLinkRequestsSingle(a.Ctx, addr, time.Time(startDate), limit)
		if err != nil {
			a.AddError(fmt.Errorf("error getting address memo link requests for address; %w", err))
			return
		}
		addrLinkRequestParents, err := memo.GetAddrLinkRequestParentsSingle(a.Ctx, addr, time.Time(startDate), limit)
		if err != nil {
			a.AddError(fmt.Errorf("error getting address memo link request parents for address; %w", err))
			return
		}
		a.Mutex.Lock()
		for _, i := range profileIndexMap[addr] {
			linkRequests := mergeLinkRequests(addrLinkRequests, addrLinkRequestParents, limit)
			a.Profiles[i].Links = append(a.Profiles[i].Links, linkRequests...)
			allLinkRequests = append(allLinkRequests, linkRequests...)
		}
		a.Mutex.Unlock()
	}
	if err := ToMemoLinkRequests(a.Ctx, linkRequestsField.Fields, allLinkRequests); err != nil {
		a.AddError(fmt.Errorf("error attaching to link requests for memo profiles; %w", err))
		return
	}
}

type linkRequestPageItem struct {
	seen    time.Time
	request *model.LinkRequest
}

func mergeLinkRequests(addrRequests []*memo.AddrLinkRequest, parentRequests []*memo.AddrLinkRequestParent, limit uint32) []*model.LinkRequest {
	items := make([]linkRequestPageItem, 0, len(addrRequests)+len(parentRequests))
	seenTxHashes := make(map[[32]byte]struct{}, len(addrRequests)+len(parentRequests))
	for _, request := range addrRequests {
		seenTxHashes[request.TxHash] = struct{}{}
		items = append(items, linkRequestPageItem{
			seen: request.Seen,
			request: &model.LinkRequest{
				Address:       model.NewAddress(request.Addr),
				TxHash:        request.TxHash,
				ParentAddress: model.NewAddress(request.ParentAddr),
				Message:       request.Message,
			},
		})
	}
	for _, request := range parentRequests {
		if _, ok := seenTxHashes[request.TxHash]; ok {
			continue
		}
		seenTxHashes[request.TxHash] = struct{}{}
		items = append(items, linkRequestPageItem{
			seen: request.Seen,
			request: &model.LinkRequest{
				Address:       model.NewAddress(request.Addr),
				TxHash:        request.TxHash,
				ParentAddress: model.NewAddress(request.ParentAddr),
				Message:       request.Message,
			},
		})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].seen.Equal(items[j].seen) {
			return bytes.Compare(items[i].request.TxHash[:], items[j].request.TxHash[:]) > 0
		}
		return items[i].seen.After(items[j].seen)
	})
	if len(items) > int(limit) {
		items = items[:limit]
	}
	requests := make([]*model.LinkRequest, len(items))
	for i := range items {
		requests[i] = items[i].request
	}
	return requests
}

func (a *MemoProfile) AttachRooms() {
	defer a.Wait.Done()
	if !a.HasField([]string{"rooms"}) {
		return
	}
	roomsField := a.Fields.GetField("rooms")
	limit, err := unmarshalPageLimit(roomsField.Arguments, "profile rooms", memo.MaxPageSize)
	if err != nil {
		a.AddError(err)
		return
	}
	lockRoomFollows, err := memo.GetAddrRoomFollows(a.Ctx, a.getAddresses(), limit)
	if err != nil {
		a.AddError(fmt.Errorf("error getting addr room follows for profile attach; %w", err))
		return
	}
	profileIndexMap := a.getProfileIndexMap()
	var allRoomFollows []*model.RoomFollow
	a.Mutex.Lock()
	for _, lockRoomFollow := range lockRoomFollows {
		if lockRoomFollow.Unfollow {
			continue
		}
		indices, ok := profileIndexMap[lockRoomFollow.Addr]
		if !ok {
			continue
		}
		for _, i := range indices {
			roomFollow := &model.RoomFollow{
				Name:     lockRoomFollow.Room,
				Address:  model.NewAddress(lockRoomFollow.Addr),
				TxHash:   lockRoomFollow.TxHash,
				Unfollow: lockRoomFollow.Unfollow,
			}
			a.Profiles[i].Rooms = append(a.Profiles[i].Rooms, roomFollow)
			allRoomFollows = append(allRoomFollows, roomFollow)
		}
	}
	a.Mutex.Unlock()
	if err := ToMemoRoomFollows(a.Ctx, roomsField.Fields, allRoomFollows); err != nil {
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
		allSetNames = append(allSetNames, &model.SetName{Address: model.NewAddress(address)})
	}
	if err := ToMemoSetNames(a.Ctx, GetPrefixFields(a.Fields, "name"), allSetNames); err != nil {
		a.AddError(fmt.Errorf("error attaching to names for memo profiles; %w", err))
		return
	}
	profileIndexMap := a.getProfileIndexMap()
	a.Mutex.Lock()
	defer a.Mutex.Unlock()
	for _, setName := range allSetNames {
		if !SetNameExists(setName) {
			continue
		}
		for _, i := range profileIndexMap[setName.Address.Addr()] {
			a.Profiles[i].Name = setName
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
		allSetProfiles = append(allSetProfiles, &model.SetProfile{Address: model.NewAddress(address)})
	}
	if err := ToMemoSetProfiles(a.Ctx, GetPrefixFields(a.Fields, "profile"), allSetProfiles); err != nil {
		a.AddError(fmt.Errorf("error attaching to profiles for memo profiles; %w", err))
		return
	}
	profileIndexMap := a.getProfileIndexMap()
	a.Mutex.Lock()
	defer a.Mutex.Unlock()
	for _, setProfile := range allSetProfiles {
		if !SetProfileExists(setProfile) {
			continue
		}
		for _, i := range profileIndexMap[setProfile.Address.Addr()] {
			a.Profiles[i].Profile = setProfile
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
		allSetPics = append(allSetPics, &model.SetPic{Address: model.NewAddress(address)})
	}
	if err := ToMemoSetPics(a.Ctx, GetPrefixFields(a.Fields, "pic"), allSetPics); err != nil {
		a.AddError(fmt.Errorf("error attaching to pics for memo profiles; %w", err))
		return
	}
	profileIndexMap := a.getProfileIndexMap()
	a.Mutex.Lock()
	defer a.Mutex.Unlock()
	for _, setPic := range allSetPics {
		if !SetPicExists(setPic) {
			continue
		}
		for _, i := range profileIndexMap[setPic.Address.Addr()] {
			a.Profiles[i].Pic = setPic
		}
	}
}

// GetMemoProfiles returns one profile per requested address, in request
// order, with fields attached. A P2SH32 address is looked up through its
// alias and reported under the address asked for.
func GetMemoProfiles(ctx context.Context, fields Fields, addresses []model.Address) ([]*model.Profile, error) {
	memoAddrs, err := ResolveMemoAddresses(ctx, addresses)
	if err != nil {
		return nil, fmt.Errorf("error resolving memo addresses for profiles; %w", err)
	}
	return GetMemoProfilesForEntries(ctx, fields, memoAddrs.Entries)
}

// GetMemoProfilesForEntries returns one profile per entry, in order. Memo
// activity is keyed by the 20-byte form, so the memo fields are attached
// once per distinct key (attaching the same key twice would read and append
// its list rows twice) and every entry with that key gets its own copy
// sharing that data. The lock is the entry's own: it is attached per entry
// under the requested address, so lock.address is the form asked for and
// lock.txs the history of that address (merged across both keys for a P2SH
// address). An entry with no key (an unlinked P2SH32 address) has no memo
// activity and gets an empty profile, without the memo topics being read
// through a zero key; its lock is still attached, since an unspent address
// can have received outputs.
func GetMemoProfilesForEntries(ctx context.Context, fields Fields, entries []MemoAddress) ([]*model.Profile, error) {
	var byKey = make(map[[25]byte]*model.Profile)
	var keyProfiles []*model.Profile
	for _, entry := range entries {
		if _, ok := byKey[entry.Key]; entry.HasKey && !ok {
			profile := &model.Profile{Address: model.NewAddress(entry.Key)}
			byKey[entry.Key] = profile
			keyProfiles = append(keyProfiles, profile)
		}
	}
	if err := ToMemoProfiles(ctx, fields.Without("lock"), keyProfiles); err != nil {
		return nil, fmt.Errorf("error attaching to memo profiles; %w", err)
	}
	var profiles = make([]*model.Profile, len(entries))
	var locks []*model.Lock
	for i, entry := range entries {
		var profile model.Profile
		if entry.HasKey {
			// A copy per requested form; the attached memo data is shared
			// and only read from here on.
			profile = *byKey[entry.Key]
		}
		profile.Address = entry.Address
		if fields.HasField("lock") {
			profile.Lock = &model.Lock{Address: entry.Address}
			locks = append(locks, profile.Lock)
		}
		profiles[i] = &profile
	}
	if len(locks) > 0 {
		if err := ToLocks(ctx, GetPrefixFields(fields, "lock."), locks); err != nil {
			return nil, fmt.Errorf("error attaching to locks for memo profiles; %w", err)
		}
	}
	return profiles, nil
}
