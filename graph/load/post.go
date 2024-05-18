package load

import (
	"errors"
	"fmt"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/memo"
	"github.com/memocash/index/graph/dataloader"
	"github.com/memocash/index/graph/model"
	"github.com/memocash/index/ref/bitcoin/wallet"
)

const postNotFoundErrorMessage = "error post not found in loader"

var postNotFoundError = fmt.Errorf(postNotFoundErrorMessage)

func IsPostNotFoundError(err error) bool {
	return errors.Is(err, postNotFoundError)
}

var Post = dataloader.NewPostLoader(dataloader.PostLoaderConfig{
	Wait: defaultWait,
	Fetch: func(txHashStrings []string) ([]*model.Post, []error) {
		var posts = make([]*model.Post, len(txHashStrings))
		var errors = make([]error, len(txHashStrings))
		for i, txHashString := range txHashStrings {
			txHash, err := chainhash.NewHashFromStr(txHashString)
			if err != nil {
				errors[i] = fmt.Errorf("error getting tx hash from string; %w", err)
				continue
			}
			memoPost, err := memo.GetPost(*txHash)
			if err != nil && !client.IsEntryNotFoundError(err) {
				errors[i] = fmt.Errorf("error getting lock memo post; %w", err)
				continue
			}
			if memoPost == nil {
				errors[i] = fmt.Errorf("error post not found: %s; %w", txHashString, postNotFoundError)
				continue
			}
			posts[i] = &model.Post{
				TxHash:  txHash.String(),
				Address: wallet.Addr(memoPost.Addr).String(),
				Text:    memoPost.Post,
			}
		}
		return posts, errors
	},
})
