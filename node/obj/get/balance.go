package get

import (
	"bytes"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/ref/bitcoin/tx/hs"
	"github.com/memocash/index/ref/bitcoin/tx/script"
	"github.com/memocash/index/ref/bitcoin/wallet"
)

type Utxo struct {
	Item item.LockUtxo

	TxLost *item.TxLost
}

type Balance struct {
	LockScript []byte
	Address    string
	Balance    int64
	Spendable  int64
	UtxoCount  int
	Spends     int
	OutUtxos   bool

	Utxos []*Utxo
}

func (b *Balance) GetBalance() error {
	lockHash := script.GetLockHash(b.LockScript)
	balance, err := item.GetLockBalance(lockHash)
	if err != nil && !client.IsEntryNotFoundError(err) {
		return jerr.Get("error getting lock balance", err)
	}
	if balance != nil {
		b.Balance = balance.Balance
		b.UtxoCount = balance.UtxoCount
		b.Spendable = balance.Spendable
		b.Spends = balance.Spends
		return nil
	}
	if err := b.GetBalanceByUtxos(); err != nil {
		return jerr.Get("error getting utxos for balance", err)
	}
	return nil
}

func (b *Balance) GetBalanceByUtxos() error {
	if err := b.attachUtxos(); err != nil {
		return jerr.Get("error attaching utxos", err)
	}
	if err := b.attachTxLosts(); err != nil {
		return jerr.Get("error attaching tx losts", err)
	}
	if err := b.CalculateWithUtxos(); err != nil {
		return jerr.Get("error calculating balance with utxos", err)
	}
	return nil
}

func (b *Balance) attachUtxos() error {
	lockHash := script.GetLockHash(b.LockScript)
	var lastUid []byte
	for {
		lockUtxos, err := item.GetLockUtxos(lockHash, lastUid)
		if err != nil {
			return jerr.Get("error getting lock utxos", err)
		}
		for _, lockUtxo := range lockUtxos {
			if bytes.Equal(lockUtxo.GetUid(), lastUid) {
				continue
			}
			b.Utxos = append(b.Utxos, &Utxo{Item: *lockUtxo})
		}
		if len(lockUtxos) < client.DefaultLimit {
			break
		}
		lastUid = lockUtxos[len(lockUtxos)-1].GetUid()
	}
	return nil
}

func (b *Balance) attachTxLosts() error {
	var txHashes = make([][]byte, len(b.Utxos))
	for i := range b.Utxos {
		txHashes[i] = b.Utxos[i].Item.Hash
	}
	txLosts, err := item.GetTxLosts(txHashes)
	if err != nil {
		return jerr.Get("error getting tx losts for utxos", err)
	}
	for _, utxo := range b.Utxos {
		for _, txLost := range txLosts {
			if bytes.Equal(txLost.TxHash, utxo.Item.Hash) {
				utxo.TxLost = txLost
				break
			}
		}
	}
	return nil
}

func (b *Balance) CalculateWithUtxos() error {
	var utxoValue int64
	var utxoSpendableValue int64
	var utxoSpendableCount int
	for _, utxo := range b.Utxos {
		if utxo.TxLost != nil {
			continue
		}
		utxoValue += utxo.Item.Value
		if !utxo.Item.Special {
			utxoSpendableValue += utxo.Item.Value
			utxoSpendableCount++
		}
		if b.OutUtxos {
			jlog.Logf("UTXO: %s:%d - %s\n", hs.GetTxString(utxo.Item.Hash), utxo.Item.Index, utxo.Item.Value)
		}
	}
	b.Balance = utxoValue
	b.UtxoCount = len(b.Utxos)
	b.Spendable = utxoSpendableValue
	b.Spends = utxoSpendableCount
	lockHash := script.GetLockHash(b.LockScript)
	var lockBalance = &item.LockBalance{
		LockHash:  lockHash,
		Balance:   b.Balance,
		Spendable: b.Spendable,
		UtxoCount: b.UtxoCount,
		Spends:    b.Spends,
	}
	err := item.Save([]item.Object{
		lockBalance,
	})
	if err != nil {
		return jerr.Get("error saving lock balance", err)
	}
	return nil
}

func NewBalanceFromAddress(address string) (*Balance, error) {
	addr := wallet.GetAddressFromString(address)
	var lockScript []byte
	var err error
	if !addr.IsSet() {
		return nil, jerr.New("error parsing address")
	} else if addr.IsP2PKH() {
		s := script.P2pkh{PkHash: addr.GetPkHash()}
		lockScript, err = s.Get()
		if err != nil {
			return nil, jerr.Get("error getting lock script for p2pkh address", err)
		}
	} else if addr.IsP2SH() {
		s := script.P2sh{ScriptHash: addr.ScriptAddress()}
		lockScript, err = s.Get()
		if err != nil {
			return nil, jerr.Get("error getting lock script for p2sh address", err)
		}
	} else {
		return nil, jerr.Newf("error unknown address type")
	}
	return &Balance{
		LockScript: lockScript,
		Address:    addr.GetEncoded(),
	}, nil
}

func NewBalance(lockScript []byte) *Balance {
	return &Balance{
		LockScript: lockScript,
	}
}
