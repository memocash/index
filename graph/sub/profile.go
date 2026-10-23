package sub

import (
	"context"
	"fmt"
	"log"

	"github.com/memocash/index/db/item/memo"
	"github.com/memocash/index/graph/attach"
	"github.com/memocash/index/graph/model"
	"github.com/memocash/index/ref/bitcoin/wallet"
)

// Profile streams profile updates for a set of addresses. Memo activity is
// keyed by the 20-byte form, so a P2SH32 address is followed through its
// P2SH20 alias and reported under the address the client asked for (under
// each, if both forms were asked for). A P2SH32 address whose script has not
// been spent yet has no alias; its link is watched and, once the first
// spend writes one, the alias is followed from then on.
type Profile struct {
	Addresses        [][25]byte
	AddrUpdateChan   chan [25]byte
	Cancel           context.CancelFunc
	Fields           attach.Fields
	MemoAddrs        *attach.MemoAddresses
	NameAddresses    [][25]byte
	ProfileAddresses [][25]byte
	PicAddresses     [][25]byte
}

func (p *Profile) Listen(ctx context.Context, addresses []model.Address, fields attach.Fields) (<-chan *model.Profile, error) {
	ctx, p.Cancel = context.WithCancel(ctx)
	p.Fields = fields
	// Watch links before resolving, so a first spend between the two is
	// still seen (by the watch rather than the lookup).
	watch, err := newLinkWatch(ctx, p2sh32Addrs(model.AddressesToLockAddrs(addresses)))
	if err != nil {
		p.Cancel()
		return nil, fmt.Errorf("error watching p2sh links for profile subscription; %w", err)
	}
	memoAddrs, err := attach.ResolveMemoAddresses(ctx, addresses)
	if err != nil {
		p.Cancel()
		return nil, fmt.Errorf("error resolving memo addresses for profile subscription; %w", err)
	}
	p.MemoAddrs = memoAddrs
	for _, addr := range memoAddrs.Keys {
		watch.Follow(wallet.Addr(addr).LockAddr())
	}
	if err := p.SetupAddresses(memoAddrs.Keys); err != nil {
		return nil, fmt.Errorf("error setting up lock hashes for profile; %w", err)
	}
	// With no resolved address yet (only unlinked P2SH32 addresses) there
	// is nothing to listen to until an alias appears; the listeners are
	// attached then, by listenAlias.
	if fields.HasField("following") && len(p.Addresses) > 0 {
		if err := p.ListenFollowing(ctx, p.Addresses); err != nil {
			return nil, fmt.Errorf("error listening following; %w", err)
		}
		if err := p.SetupFollowingLockHashes(ctx, fields); err != nil {
			return nil, fmt.Errorf("error setting up following lock hashes for profile subscription; %w", err)
		}
	}
	if fields.HasField("followers") && len(p.Addresses) > 0 {
		if err := p.ListenFollowers(ctx, p.Addresses); err != nil {
			return nil, fmt.Errorf("error listening followers; %w", err)
		}
		if err := p.SetupFollowersLockHashes(ctx, fields); err != nil {
			return nil, fmt.Errorf("error setting up followers lock hashes for profile subscription; %w", err)
		}
	}
	if fields.HasField("name") {
		p.NameAddresses = append(p.NameAddresses, p.Addresses...)
	}
	if len(p.NameAddresses) > 0 {
		if err := p.ListenNames(ctx, p.NameAddresses); err != nil {
			return nil, fmt.Errorf("error listening names; %w", err)
		}
	}
	if fields.HasField("profile") {
		p.ProfileAddresses = append(p.ProfileAddresses, p.Addresses...)
	}
	if len(p.ProfileAddresses) > 0 {
		if err := p.ListenProfiles(ctx, p.ProfileAddresses); err != nil {
			return nil, fmt.Errorf("error listening profiles; %w", err)
		}
	}
	if fields.HasField("pic") {
		p.PicAddresses = append(p.PicAddresses, p.Addresses...)
	}
	if len(p.PicAddresses) > 0 {
		if err := p.ListenPics(ctx, p.PicAddresses); err != nil {
			return nil, fmt.Errorf("error listening pics; %w", err)
		}
	}
	if watch.Aliases != nil {
		go func() {
			// The watch ends only with the subscription or with its link
			// stream; either way the subscription is over.
			defer p.Cancel()
			for link := range watch.Aliases {
				key := link.Other.Addr()
				p.MemoAddrs.AddAlias(key, link.Addr)
				if !watch.Follow(link.Other) {
					// The alias was asked for directly and is already
					// followed; the new form just gets a snapshot.
					if !p.forward(ctx, key) {
						return
					}
					continue
				}
				if err := p.listenAlias(ctx, key); err != nil {
					log.Printf("error following new p2sh alias for profile subscription; %v", err)
					return
				}
			}
		}()
	}
	return p.GetProfileChan(ctx), nil
}

// listenAlias attaches the requested listeners for an alias that appeared
// after the start, then emits its current profile. The listeners are
// registered before the profile is read, so a memo action written at any
// point after the link (the first spend's own actions are saved right after
// its link) is reflected either in that read or in a later update.
func (p *Profile) listenAlias(ctx context.Context, alias [25]byte) error {
	addrs := [][25]byte{alias}
	if p.Fields.HasField("following") {
		if err := p.ListenFollowing(ctx, addrs); err != nil {
			return err
		}
	}
	if p.Fields.HasField("followers") {
		if err := p.ListenFollowers(ctx, addrs); err != nil {
			return err
		}
	}
	if p.Fields.HasField("name") {
		if err := p.ListenNames(ctx, addrs); err != nil {
			return err
		}
	}
	if p.Fields.HasField("profile") {
		if err := p.ListenProfiles(ctx, addrs); err != nil {
			return err
		}
	}
	if p.Fields.HasField("pic") {
		if err := p.ListenPics(ctx, addrs); err != nil {
			return err
		}
	}
	select {
	case <-ctx.Done():
	case p.AddrUpdateChan <- alias:
	}
	return nil
}

func (p *Profile) SetupAddresses(addresses [][25]byte) error {
	p.Addresses = addresses
	p.AddrUpdateChan = make(chan [25]byte)
	return nil
}

func (p *Profile) SetupFollowingLockHashes(ctx context.Context, fields attach.Fields) error {
	if !fields.HasFieldAny([]string{
		"following.follow_lock.profile.name",
		"following.follow_lock.profile.profile",
		"following.follow_lock.profile.pic",
	}) {
		return nil
	}
	lockMemoFollows, err := memo.GetAddrFollows(ctx, p.Addresses)
	if err != nil {
		return fmt.Errorf("error getting lock memo follows for profile lock hashes; %w", err)
	}
	var lastMemoFollow *memo.AddrFollow
	for _, lockMemoFollow := range lockMemoFollows {
		if lastMemoFollow != nil && lockMemoFollow.FollowAddr == lastMemoFollow.FollowAddr {
			continue
		}
		if !lockMemoFollow.Unfollow {
			if fields.HasField("following.follow_lock.profile.name") {
				p.NameAddresses = append(p.NameAddresses, lockMemoFollow.FollowAddr)
			}
			if fields.HasField("following.follow_lock.profile.profile") {
				p.ProfileAddresses = append(p.ProfileAddresses, lockMemoFollow.FollowAddr)
			}
			if fields.HasField("following.follow_lock.profile.pic") {
				p.PicAddresses = append(p.PicAddresses, lockMemoFollow.FollowAddr)
			}
		}
		lastMemoFollow = lockMemoFollow
	}
	return nil
}

func (p *Profile) SetupFollowersLockHashes(ctx context.Context, fields attach.Fields) error {
	if !fields.HasFieldAny([]string{
		"followers.lock.profile.name",
		"followers.lock.profile.profile",
		"followers.lock.profile.pic",
	}) {
		return nil
	}
	lockMemoFolloweds, err := memo.GetAddrFolloweds(ctx, p.Addresses)
	if err != nil {
		return fmt.Errorf("error getting lock memo followeds for profile lock hashes; %w", err)
	}
	var lastLockMemoFollowed *memo.AddrFollowed
	for _, lockMemoFollowed := range lockMemoFolloweds {
		if lastLockMemoFollowed != nil && lockMemoFollowed.Addr == lastLockMemoFollowed.Addr {
			continue
		}
		if !lockMemoFollowed.Unfollow {
			if fields.HasField("followers.lock.profile.name") {
				p.NameAddresses = append(p.NameAddresses, lockMemoFollowed.Addr)
			}
			if fields.HasField("followers.lock.profile.profile") {
				p.ProfileAddresses = append(p.ProfileAddresses, lockMemoFollowed.Addr)
			}
			if fields.HasField("followers.lock.profile.pic") {
				p.PicAddresses = append(p.PicAddresses, lockMemoFollowed.Addr)
			}
		}
		lastLockMemoFollowed = lockMemoFollowed
	}
	return nil
}

// forward sends an updated address to the profile loop until the
// subscription ends.
func (p *Profile) forward(ctx context.Context, addr [25]byte) bool {
	select {
	case <-ctx.Done():
		return false
	case p.AddrUpdateChan <- addr:
		return true
	}
}

func (p *Profile) ListenFollowing(ctx context.Context, addrs [][25]byte) error {
	addrMemoFollowingListener, err := memo.ListenAddrFollows(ctx, addrs)
	if err != nil {
		p.Cancel()
		return fmt.Errorf("error getting lock memo following listener for profile subscription; %w", err)
	}
	go func() {
		defer p.Cancel()
		for {
			select {
			case <-ctx.Done():
				return
			case addrMemoFollow, ok := <-addrMemoFollowingListener:
				if !ok || !p.forward(ctx, addrMemoFollow.Addr) {
					return
				}
			}
		}
	}()
	return nil
}

func (p *Profile) ListenFollowers(ctx context.Context, addrs [][25]byte) error {
	lockMemoFollowerListener, err := memo.ListenAddrFolloweds(ctx, addrs)
	if err != nil {
		p.Cancel()
		return fmt.Errorf("error getting memo followers listener for profile subscription; %w", err)
	}
	go func() {
		defer p.Cancel()
		for {
			select {
			case <-ctx.Done():
				return
			case lockMemoFollow, ok := <-lockMemoFollowerListener:
				if !ok || !p.forward(ctx, lockMemoFollow.Addr) {
					return
				}
			}
		}
	}()
	return nil
}

func (p *Profile) ListenNames(ctx context.Context, addrs [][25]byte) error {
	lockMemoNameListener, err := memo.ListenAddrNames(ctx, addrs)
	if err != nil {
		p.Cancel()
		return fmt.Errorf("error getting addr memo name listener for profile subscription; %w", err)
	}
	go func() {
		defer p.Cancel()
		for {
			select {
			case <-ctx.Done():
				return
			case lockMemoName, ok := <-lockMemoNameListener:
				if !ok || !p.forward(ctx, lockMemoName.Addr) {
					return
				}
			}
		}
	}()
	return nil
}

func (p *Profile) ListenProfiles(ctx context.Context, addrs [][25]byte) error {
	lockMemoProfileListener, err := memo.ListenAddrProfiles(ctx, addrs)
	if err != nil {
		p.Cancel()
		return fmt.Errorf("error getting addr memo profile listener for profile subscription; %w", err)
	}
	go func() {
		defer p.Cancel()
		for {
			select {
			case <-ctx.Done():
				return
			case lockMemoProfile, ok := <-lockMemoProfileListener:
				if !ok || !p.forward(ctx, lockMemoProfile.Addr) {
					return
				}
			}
		}
	}()
	return nil
}

func (p *Profile) ListenPics(ctx context.Context, addrs [][25]byte) error {
	lockMemoProfilePicListener, err := memo.ListenAddrProfilePics(ctx, addrs)
	if err != nil {
		p.Cancel()
		return fmt.Errorf("error getting addr memo profile pic listener for profile subscription; %w", err)
	}
	go func() {
		defer p.Cancel()
		for {
			select {
			case <-ctx.Done():
				return
			case lockMemoProfilePic, ok := <-lockMemoProfilePicListener:
				if !ok || !p.forward(ctx, lockMemoProfilePic.Addr) {
					return
				}
			}
		}
	}()
	return nil
}

// GetProfileChan attaches and emits a profile for each updated address. The
// update channel is never closed: the listeners that send on it stop with
// the context, and a close would race their sends.
func (p *Profile) GetProfileChan(ctx context.Context) <-chan *model.Profile {
	var profileChan = make(chan *model.Profile)
	go func() {
		defer func() {
			close(profileChan)
			p.Cancel()
		}()
		for {
			select {
			case <-ctx.Done():
				return
			case addr := <-p.AddrUpdateChan:
				// Attached under the memo key, reported under each address
				// the client asked for it by.
				profiles, err := attach.GetMemoProfilesForEntries(ctx, p.Fields, p.MemoAddrs.KeyEntries(addr))
				if err != nil {
					log.Printf("error attaching to memo profiles for profile subscription resolver; %v", err)
					return
				}
				for _, profile := range profiles {
					select {
					case <-ctx.Done():
						return
					case profileChan <- profile:
					}
				}
			}
		}
	}()
	return profileChan
}
