package sub

import (
	"context"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/admin/graph/load"
	"github.com/memocash/index/admin/graph/model"
	"github.com/memocash/index/db/item/memo"
	"github.com/memocash/index/ref/bitcoin/wallet"
)

type Profile struct {
	Addresses        [][25]byte
	AddrUpdateChan   chan [25]byte
	Cancel           context.CancelFunc
	NameAddresses    [][25]byte
	ProfileAddresses [][25]byte
	PicAddresses     [][25]byte
}

func (p *Profile) Listen(ctx context.Context, addresses []string, fields load.Fields) (<-chan *model.Profile, error) {
	ctx, p.Cancel = context.WithCancel(ctx)
	if err := p.SetupAddresses(addresses); err != nil {
		return nil, jerr.Get("error setting up lock hashes for profile", err)
	}
	if fields.HasField("following") {
		if err := p.ListenFollowing(ctx, p.Addresses); err != nil {
			return nil, jerr.Get("error listening following", err)
		}
		if err := p.SetupFollowingLockHashes(ctx, fields); err != nil {
			return nil, jerr.Get("error setting up following lock hashes for profile subscription", err)
		}
	}
	if fields.HasField("followers") {
		if err := p.ListenFollowers(ctx, p.Addresses); err != nil {
			return nil, jerr.Get("error listening followers", err)
		}
		if err := p.SetupFollowersLockHashes(ctx, fields); err != nil {
			return nil, jerr.Get("error setting up followers lock hashes for profile subscription", err)
		}
	}
	if fields.HasField("name") {
		p.NameAddresses = append(p.NameAddresses, p.Addresses...)
	}
	if len(p.NameAddresses) > 0 {
		if err := p.ListenNames(ctx, p.NameAddresses); err != nil {
			return nil, jerr.Get("error listening names", err)
		}
	}
	if fields.HasField("profile") {
		p.ProfileAddresses = append(p.ProfileAddresses, p.Addresses...)
	}
	if len(p.ProfileAddresses) > 0 {
		if err := p.ListenProfiles(ctx, p.ProfileAddresses); err != nil {
			return nil, jerr.Get("error listening profiles", err)
		}
	}
	if fields.HasField("pic") {
		p.PicAddresses = append(p.PicAddresses, p.Addresses...)
	}
	if len(p.PicAddresses) > 0 {
		if err := p.ListenPics(ctx, p.PicAddresses); err != nil {
			return nil, jerr.Get("error listening pics", err)
		}
	}
	return p.GetProfileChan(ctx), nil
}

func (p *Profile) SetupAddresses(addresses []string) error {
	for _, address := range addresses {
		addr, err := wallet.GetAddrFromString(address)
		if err != nil {
			return jerr.Get("error getting address from string for profile", err)
		}
		p.Addresses = append(p.Addresses, *addr)
	}
	p.AddrUpdateChan = make(chan [25]byte)
	return nil
}

func (p *Profile) SetupFollowingLockHashes(ctx context.Context, fields load.Fields) error {
	if !fields.HasFieldAny([]string{
		"following.follow_lock.profile.name",
		"following.follow_lock.profile.profile",
		"following.follow_lock.profile.pic",
	}) {
		return nil
	}
	lockMemoFollows, err := memo.GetAddrFollows(ctx, p.Addresses)
	if err != nil {
		return jerr.Get("error getting lock memo follows for profile lock hashes", err)
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

func (p *Profile) SetupFollowersLockHashes(ctx context.Context, fields load.Fields) error {
	if !fields.HasFieldAny([]string{
		"followers.lock.profile.name",
		"followers.lock.profile.profile",
		"followers.lock.profile.pic",
	}) {
		return nil
	}
	lockMemoFolloweds, err := memo.GetAddrFolloweds(ctx, p.Addresses)
	if err != nil {
		return jerr.Get("error getting lock memo followeds for profile lock hashes", err)
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

func (p *Profile) ListenFollowing(ctx context.Context, addrs [][25]byte) error {
	addrMemoFollowingListener, err := memo.ListenAddrFollows(ctx, addrs)
	if err != nil {
		p.Cancel()
		return jerr.Get("error getting lock memo following listener for profile subscription", err)
	}
	go func() {
		defer p.Cancel()
		for {
			select {
			case <-ctx.Done():
				return
			case addrMemoFollow, ok := <-addrMemoFollowingListener:
				if !ok {
					return
				}
				p.AddrUpdateChan <- addrMemoFollow.Addr
			}
		}
	}()
	return nil
}

func (p *Profile) ListenFollowers(ctx context.Context, addrs [][25]byte) error {
	lockMemoFollowerListener, err := memo.ListenAddrFolloweds(ctx, addrs)
	if err != nil {
		p.Cancel()
		return jerr.Get("error getting memo followers listener for profile subscription", err)
	}
	go func() {
		defer p.Cancel()
		for {
			select {
			case <-ctx.Done():
				return
			case lockMemoFollow, ok := <-lockMemoFollowerListener:
				if !ok {
					return
				}
				p.AddrUpdateChan <- lockMemoFollow.Addr
			}
		}
	}()
	return nil
}

func (p *Profile) ListenNames(ctx context.Context, addrs [][25]byte) error {
	lockMemoNameListener, err := memo.ListenAddrNames(ctx, addrs)
	if err != nil {
		p.Cancel()
		return jerr.Get("error getting addr memo name listener for profile subscription", err)
	}
	go func() {
		defer p.Cancel()
		for {
			select {
			case <-ctx.Done():
				return
			case lockMemoName, ok := <-lockMemoNameListener:
				if !ok {
					return
				}
				p.AddrUpdateChan <- lockMemoName.Addr
			}
		}
	}()
	return nil
}

func (p *Profile) ListenProfiles(ctx context.Context, addrs [][25]byte) error {
	lockMemoProfileListener, err := memo.ListenAddrProfiles(ctx, addrs)
	if err != nil {
		p.Cancel()
		return jerr.Get("error getting addr memo profile listener for profile subscription", err)
	}
	go func() {
		defer p.Cancel()
		for {
			select {
			case <-ctx.Done():
				return
			case lockMemoProfile, ok := <-lockMemoProfileListener:
				if !ok {
					return
				}
				p.AddrUpdateChan <- lockMemoProfile.Addr
			}
		}
	}()
	return nil
}

func (p *Profile) ListenPics(ctx context.Context, addrs [][25]byte) error {
	lockMemoProfilePicListener, err := memo.ListenAddrProfilePics(ctx, addrs)
	if err != nil {
		p.Cancel()
		return jerr.Get("error getting addr memo profile pic listener for profile subscription", err)
	}
	go func() {
		defer p.Cancel()
		for {
			select {
			case <-ctx.Done():
				return
			case lockMemoProfilePic, ok := <-lockMemoProfilePicListener:
				if !ok {
					return
				}
				p.AddrUpdateChan <- lockMemoProfilePic.Addr
			}
		}
	}()
	return nil
}

func (p *Profile) GetProfileChan(ctx context.Context) <-chan *model.Profile {
	var profileChan = make(chan *model.Profile)
	go func() {
		defer func() {
			close(p.AddrUpdateChan)
			close(profileChan)
			p.Cancel()
		}()
		for {
			select {
			case <-ctx.Done():
				return
			case addr, ok := <-p.AddrUpdateChan:
				if !ok {
					return
				}
				profile, err := load.GetProfile(ctx, wallet.Addr(addr).String())
				if err != nil {
					jerr.Get("error getting profile from dataloader for profile subscription resolver", err).Print()
					return
				}
				profileChan <- profile
			}
		}
	}()
	return profileChan
}
