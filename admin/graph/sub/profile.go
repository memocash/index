package sub

import (
	"bytes"
	"context"
	"encoding/hex"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/admin/graph/dataloader"
	"github.com/memocash/index/admin/graph/load"
	"github.com/memocash/index/admin/graph/model"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/node/obj/get"
	"github.com/memocash/index/ref/bitcoin/tx/script"
	"github.com/memocash/index/ref/bitcoin/wallet"
)

type Profile struct {
	LockHashes         [][]byte
	LockHashUpdateChan chan []byte
	LockHashAddressMap map[string]string
	NeedsLockHashMap   [][]byte
	Cancel             context.CancelFunc
	NameLockHashes     [][]byte
	ProfileLockHashes  [][]byte
	PicLockHashes      [][]byte
}

func (p *Profile) Listen(ctx context.Context, addresses, preloads []string) (<-chan *model.Profile, error) {
	ctx, p.Cancel = context.WithCancel(ctx)
	if err := p.SetupLockHashes(addresses); err != nil {
		return nil, jerr.Get("error setting up lock hashes for profile", err)
	}
	if jutil.StringInSlice("following", preloads) {
		if err := p.ListenFollowing(ctx, p.LockHashes); err != nil {
			return nil, jerr.Get("error listening following", err)
		}
		if err := p.SetupFollowingLockHashes(ctx, preloads); err != nil {
			return nil, jerr.Get("error setting up following lock hashes for profile subscription", err)
		}
	}
	if jutil.StringInSlice("followers", preloads) {
		if err := p.ListenFollowers(ctx, p.LockHashes); err != nil {
			return nil, jerr.Get("error listening followers", err)
		}
		if err := p.SetupFollowersLockHashes(ctx, preloads); err != nil {
			return nil, jerr.Get("error setting up followers lock hashes for profile subscription", err)
		}
	}
	if err := p.SetupNeedsLockHashMap(); err != nil {
		return nil, jerr.Get("error setting up needs lock hash map for profile sub", err)
	}
	if jutil.StringInSlice("name", preloads) {
		p.NameLockHashes = append(p.NameLockHashes, p.LockHashes...)
	}
	if len(p.NameLockHashes) > 0 {
		if err := p.ListenNames(ctx, p.NameLockHashes); err != nil {
			return nil, jerr.Get("error listening names", err)
		}
	}
	if jutil.StringInSlice("profile", preloads) {
		p.ProfileLockHashes = append(p.ProfileLockHashes, p.LockHashes...)
	}
	if len(p.ProfileLockHashes) > 0 {
		if err := p.ListenProfiles(ctx, p.ProfileLockHashes); err != nil {
			return nil, jerr.Get("error listening profiles", err)
		}
	}
	if jutil.StringInSlice("pic", preloads) {
		p.PicLockHashes = append(p.PicLockHashes, p.LockHashes...)
	}
	if len(p.PicLockHashes) > 0 {
		if err := p.ListenPics(ctx, p.PicLockHashes); err != nil {
			return nil, jerr.Get("error listening pics", err)
		}
	}
	return p.GetProfileChan(ctx), nil
}

func (p *Profile) SetupLockHashes(addresses []string) error {
	p.LockHashAddressMap = make(map[string]string)
	for _, address := range addresses {
		lockScript, err := get.LockScriptFromAddress(wallet.GetAddressFromString(address))
		if err != nil {
			return jerr.Get("error getting lock script for profile subscription", err)
		}
		lockHash := script.GetLockHash(lockScript)
		p.LockHashes = append(p.LockHashes, lockHash)
		p.LockHashAddressMap[hex.EncodeToString(lockHash)] = wallet.GetAddressStringFromPkScript(lockScript)
	}
	p.LockHashUpdateChan = make(chan []byte)
	return nil
}

func (p *Profile) SetupFollowingLockHashes(ctx context.Context, preloads []string) error {
	if !jutil.StringsInSlice([]string{
		"following.follow_lock.profile.name",
		"following.follow_lock.profile.profile",
		"following.follow_lock.profile.pic",
	}, preloads) {
		return nil
	}
	memoFollows, err := item.GetMemoFollow(ctx, p.LockHashes)
	if err != nil {
		return jerr.Get("error getting memo follows for profile lock hashes", err)
	}
	var lastMemoFollow *item.MemoFollow
	for _, memoFollow := range memoFollows {
		if lastMemoFollow != nil && bytes.Equal(memoFollow.Follow, lastMemoFollow.Follow) {
			continue
		}
		if !memoFollow.Unfollow {
			if jutil.StringInSlice("following.follow_lock.profile.name", preloads) {
				p.NameLockHashes = append(p.NameLockHashes, memoFollow.Follow)
			}
			if jutil.StringInSlice("following.follow_lock.profile.profile", preloads) {
				p.ProfileLockHashes = append(p.ProfileLockHashes, memoFollow.Follow)
			}
			if jutil.StringInSlice("following.follow_lock.profile.pic", preloads) {
				p.PicLockHashes = append(p.PicLockHashes, memoFollow.Follow)
			}
			if _, ok := p.LockHashAddressMap[hex.EncodeToString(memoFollow.Follow)]; !ok &&
				!jutil.InByteArray(memoFollow.Follow, p.NeedsLockHashMap) {
				p.NeedsLockHashMap = append(p.NeedsLockHashMap, memoFollow.Follow)
			}
		}
		lastMemoFollow = memoFollow
	}
	return nil
}

func (p *Profile) SetupFollowersLockHashes(ctx context.Context, preloads []string) error {
	if !jutil.StringsInSlice([]string{
		"followers.lock.profile.name",
		"followers.lock.profile.profile",
		"followers.lock.profile.pic",
	}, preloads) {
		return nil
	}
	memoFolloweds, err := item.GetMemoFollowed(ctx, p.LockHashes)
	if err != nil {
		return jerr.Get("error getting memo followeds for profile lock hashes", err)
	}
	var lastMemoFollowed *item.MemoFollowed
	for _, memoFollowed := range memoFolloweds {
		if lastMemoFollowed != nil && bytes.Equal(memoFollowed.LockHash, lastMemoFollowed.LockHash) {
			continue
		}
		if !memoFollowed.Unfollow {
			if jutil.StringInSlice("followers.lock.profile.name", preloads) {
				p.NameLockHashes = append(p.NameLockHashes, memoFollowed.LockHash)
			}
			if jutil.StringInSlice("followers.lock.profile.profile", preloads) {
				p.ProfileLockHashes = append(p.ProfileLockHashes, memoFollowed.LockHash)
			}
			if jutil.StringInSlice("followers.lock.profile.pic", preloads) {
				p.PicLockHashes = append(p.PicLockHashes, memoFollowed.LockHash)
			}
			if _, ok := p.LockHashAddressMap[hex.EncodeToString(memoFollowed.LockHash)]; !ok &&
				!jutil.InByteArray(memoFollowed.LockHash, p.NeedsLockHashMap) {
				p.NeedsLockHashMap = append(p.NeedsLockHashMap, memoFollowed.LockHash)
			}
		}
		lastMemoFollowed = memoFollowed
	}
	return nil
}

func (p *Profile) SetupNeedsLockHashMap() error {
	if len(p.NeedsLockHashMap) == 0 {
		return nil
	}
	lockAddresses, err := item.GetLockAddresses(p.NeedsLockHashMap)
	if err != nil {
		return jerr.Get("error getting lock addresses for profile following needs lock hash map", err)
	}
	for _, lockAddress := range lockAddresses {
		p.LockHashAddressMap[hex.EncodeToString(lockAddress.LockHash)] = lockAddress.Address
	}
	return nil
}

func (p *Profile) ListenFollowing(ctx context.Context, lockHashes [][]byte) error {
	memoFollowingListener, err := item.ListenMemoFollows(ctx, lockHashes)
	if err != nil {
		p.Cancel()
		return jerr.Get("error getting memo following listener for profile subscription", err)
	}
	go func() {
		defer p.Cancel()
		for {
			select {
			case <-ctx.Done():
				return
			case memoFollow, ok := <-memoFollowingListener:
				if !ok {
					return
				}
				p.LockHashUpdateChan <- memoFollow.LockHash
			}
		}
	}()
	return nil
}

func (p *Profile) ListenFollowers(ctx context.Context, lockHashes [][]byte) error {
	memoFollowerListener, err := item.ListenMemoFolloweds(ctx, lockHashes)
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
			case memoFollow, ok := <-memoFollowerListener:
				if !ok {
					return
				}
				p.LockHashUpdateChan <- memoFollow.LockHash
			}
		}
	}()
	return nil
}

func (p *Profile) ListenNames(ctx context.Context, lockHashes [][]byte) error {
	lockMemoNameListener, err := item.ListenLockMemoNames(ctx, lockHashes)
	if err != nil {
		p.Cancel()
		return jerr.Get("error getting lock memo name listener for profile subscription", err)
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
				p.LockHashUpdateChan <- lockMemoName.LockHash
			}
		}
	}()
	return nil
}

func (p *Profile) ListenProfiles(ctx context.Context, lockHashes [][]byte) error {
	memoProfileListener, err := item.ListenMemoProfiles(ctx, lockHashes)
	if err != nil {
		p.Cancel()
		return jerr.Get("error getting memo profile listener for profile subscription", err)
	}
	go func() {
		defer p.Cancel()
		for {
			select {
			case <-ctx.Done():
				return
			case memoProfile, ok := <-memoProfileListener:
				if !ok {
					return
				}
				p.LockHashUpdateChan <- memoProfile.LockHash
			}
		}
	}()
	return nil
}

func (p *Profile) ListenPics(ctx context.Context, lockHashes [][]byte) error {
	memoProfilePicListener, err := item.ListenMemoProfilePics(ctx, lockHashes)
	if err != nil {
		p.Cancel()
		return jerr.Get("error getting memo profile pic listener for profile subscription", err)
	}
	go func() {
		defer p.Cancel()
		for {
			select {
			case <-ctx.Done():
				return
			case memoProfilePic, ok := <-memoProfilePicListener:
				if !ok {
					return
				}
				p.LockHashUpdateChan <- memoProfilePic.LockHash
			}
		}
	}()
	return nil
}

func (p *Profile) GetProfileChan(ctx context.Context) <-chan *model.Profile {
	var profileChan = make(chan *model.Profile)
	go func() {
		defer func() {
			close(p.LockHashUpdateChan)
			close(profileChan)
			p.Cancel()
		}()
		for {
			select {
			case <-ctx.Done():
				return
			case lockHash, ok := <-p.LockHashUpdateChan:
				if !ok {
					return
				}
				address, ok := p.LockHashAddressMap[hex.EncodeToString(lockHash)]
				if !ok {
					jerr.New("Unable to find address for profile chan lock hash").Print()
					continue
				}
				profile, err := dataloader.NewProfileLoader(load.ProfileLoaderConfig).Load(address)
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
