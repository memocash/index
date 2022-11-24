package resolver

import (
	"context"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/admin/graph/dataloader"
	"github.com/memocash/index/admin/graph/model"
	"github.com/memocash/index/ref/bitcoin/wallet"
)

func AddrLoader(ctx context.Context, addressString string) (*model.Addr, error) {
	address, err := wallet.GetAddrFromString(addressString)
	if err != nil {
		return nil, jerr.Getf(err, "error getting address from dataloader: %s", addressString)
	}
	var addr = &model.Addr{Address: address.String()}
	if jutil.StringInSlice("balance", GetPreloads(ctx)) {
		balance, err := dataloader.NewAddressBalanceLoader(addressBalanceLoaderConfig).Load(address.String())
		if err != nil {
			return nil, jerr.Getf(err, "error getting address balance from dataloader: %s", addressString)
		}
		addr.Balance = balance
	}
	return addr, nil
}
