package load

import (
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/admin/graph/dataloader"
	"github.com/memocash/index/admin/graph/model"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/memo"
	"github.com/memocash/index/ref/bitcoin/wallet"
)

const postNotFoundErrorMessage = "error post not found in loader"

var postNotFoundError = jerr.New(postNotFoundErrorMessage)

func IsPostNotFoundError(err error) bool {
	return jerr.HasError(err, postNotFoundErrorMessage)
}

var Post = dataloader.NewPostLoader(dataloader.PostLoaderConfig{
	Wait: defaultWait,
	Fetch: func(txHashStrings []string) ([]*model.Post, []error) {
		var posts = make([]*model.Post, len(txHashStrings))
		var errors = make([]error, len(txHashStrings))
		for i, txHashString := range txHashStrings {
			txHash, err := chainhash.NewHashFromStr(txHashString)
			if err != nil {
				errors[i] = jerr.Get("error getting tx hash from string", err)
				continue
			}
			memoPost, err := memo.GetPost(*txHash)
			if err != nil && !client.IsEntryNotFoundError(err) {
				errors[i] = jerr.Get("error getting lock memo post", err)
				continue
			}
			if memoPost == nil {
				errors[i] = jerr.Getf(postNotFoundError, "error post not found: "+txHashString)
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
