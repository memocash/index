package resolver

import (
	"bytes"
	"encoding/hex"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/admin/graph/dataloader"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/ref/bitcoin/wallet"
	"time"
)

var lockAddressLoaderConfig = dataloader.LockAddressLoaderConfig{
	Wait:     2 * time.Millisecond,
	MaxBatch: 100,
	Fetch: func(lockHashStrings []string) ([]*wallet.Address, []error) {
		var addresses = make([]*wallet.Address, len(lockHashStrings))
		var errors = make([]error, len(lockHashStrings))
		var lockHashes = make([][]byte, len(lockHashStrings))
		for i := range lockHashStrings {
			lockHash, err := hex.DecodeString(lockHashStrings[i])
			if err != nil {
				errors[i] = jerr.Get("error decoding lock hash for lock hash data loader", err)
				continue
			}
			lockHashes[i] = lockHash
		}
		lockAddresses, err := item.GetLockAddresses(jutil.RemoveDupesAndEmpties(lockHashes))
		if err != nil {
			return nil, []error{jerr.Get("error getting lock address for lock hash data loader", err)}
		}
	LockHashesLoop:
		for i := range lockHashes {
			for _, lockAddress := range lockAddresses {
				if bytes.Equal(lockHashes[i], lockAddress.LockHash) {
					address, err := wallet.GetAddressFromStringErr(lockAddress.Address)
					if err != nil {
						errors[i] = jerr.Get("error getting address from string for lock address dataloader", err)
						continue LockHashesLoop
					}
					addresses[i] = address
					continue LockHashesLoop
				}
			}
			errors[i] = jerr.New("error getting lock address for lock hash data loader")
		}
		return addresses, errors
	},
}
