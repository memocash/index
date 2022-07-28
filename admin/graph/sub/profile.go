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
}

func (p *Profile) Listen(ctx context.Context, addresses, preloads []string) (<-chan *model.Profile, error) {
	var lockHashes [][]byte
	var lockHashAddressMap = make(map[string]string)
	for _, address := range addresses {
		lockScript, err := get.LockScriptFromAddress(wallet.GetAddressFromString(address))
		if err != nil {
			return nil, jerr.Get("error getting lock script for profile subscription", err)
		}
		lockHash := script.GetLockHash(lockScript)
		lockHashes = append(lockHashes, lockHash)
		lockHashAddressMap[hex.EncodeToString(lockHash)] = wallet.GetAddressStringFromPkScript(lockScript)
	}
	ctx, cancel := context.WithCancel(ctx)
	var lockHashUpdateChan = make(chan []byte)
	if jutil.StringInSlice("following", preloads) {
		memoFollowingListener, err := item.ListenMemoFollows(ctx, lockHashes)
		if err != nil {
			cancel()
			return nil, jerr.Get("error getting memo following listener for profile subscription", err)
		}
		go func() {
			defer cancel()
			for {
				select {
				case <-ctx.Done():
					return
				case memoFollow, ok := <-memoFollowingListener:
					if !ok {
						return
					}
					lockHashUpdateChan <- memoFollow.LockHash
				}
			}
		}()
	}
	if jutil.StringInSlice("followers", preloads) {
		memoFollowerListener, err := item.ListenMemoFolloweds(ctx, lockHashes)
		if err != nil {
			cancel()
			return nil, jerr.Get("error getting memo followers listener for profile subscription", err)
		}
		go func() {
			defer cancel()
			for {
				select {
				case <-ctx.Done():
					return
				case memoFollow, ok := <-memoFollowerListener:
					if !ok {
						return
					}
					lockHashUpdateChan <- memoFollow.LockHash
				}
			}
		}()
	}
	if jutil.StringInSlice("name", preloads) {
		memoNameListener, err := item.ListenMemoNames(ctx, lockHashes)
		if err != nil {
			cancel()
			return nil, jerr.Get("error getting memo name listener for profile subscription", err)
		}
		go func() {
			defer cancel()
			for {
				select {
				case <-ctx.Done():
					return
				case memoName, ok := <-memoNameListener:
					if !ok {
						return
					}
					lockHashUpdateChan <- memoName.LockHash
				}
			}
		}()
	}
	if jutil.StringInSlice("profile", preloads) {
		memoProfileListener, err := item.ListenMemoProfiles(ctx, lockHashes)
		if err != nil {
			cancel()
			return nil, jerr.Get("error getting memo profile listener for profile subscription", err)
		}
		go func() {
			defer cancel()
			for {
				select {
				case <-ctx.Done():
					return
				case memoProfile, ok := <-memoProfileListener:
					if !ok {
						return
					}
					lockHashUpdateChan <- memoProfile.LockHash
				}
			}
		}()
	}
	if jutil.StringInSlice("pic", preloads) {
		memoProfilePicListener, err := item.ListenMemoProfilePics(ctx, lockHashes)
		if err != nil {
			cancel()
			return nil, jerr.Get("error getting memo profile pic listener for profile subscription", err)
		}
		go func() {
			defer cancel()
			for {
				select {
				case <-ctx.Done():
					return
				case memoProfilePic, ok := <-memoProfilePicListener:
					if !ok {
						return
					}
					lockHashUpdateChan <- memoProfilePic.LockHash
				}
			}
		}()
	}
	var profileChan = make(chan *model.Profile)
	go func() {
		defer func() {
			close(lockHashUpdateChan)
			close(profileChan)
			cancel()
		}()
		for {
			select {
			case <-ctx.Done():
				return
			case lockHash, ok := <-lockHashUpdateChan:
				if !ok {
					return
				}
				address, ok := lockHashAddressMap[hex.EncodeToString(lockHash)]
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
	return profileChan, nil
}
