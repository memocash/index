package sub

import (
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
	Cancel             context.CancelFunc
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
	}
	if jutil.StringInSlice("followers", preloads) {
		if err := p.ListenFollowers(ctx, p.LockHashes); err != nil {
			return nil, jerr.Get("error listening followers", err)
		}
	}
	if jutil.StringInSlice("name", preloads) {
		if err := p.ListenNames(ctx, p.LockHashes); err != nil {
			return nil, jerr.Get("error listening names", err)
		}
	}
	if jutil.StringInSlice("profile", preloads) {
		if err := p.ListenProfiles(ctx, p.LockHashes); err != nil {
			return nil, jerr.Get("error listening profiles", err)
		}
	}
	if jutil.StringInSlice("pic", preloads) {
		if err := p.ListenPics(ctx, p.LockHashes); err != nil {
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
	memoNameListener, err := item.ListenMemoNames(ctx, lockHashes)
	if err != nil {
		p.Cancel()
		return jerr.Get("error getting memo name listener for profile subscription", err)
	}
	go func() {
		defer p.Cancel()
		for {
			select {
			case <-ctx.Done():
				return
			case memoName, ok := <-memoNameListener:
				if !ok {
					return
				}
				p.LockHashUpdateChan <- memoName.LockHash
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
