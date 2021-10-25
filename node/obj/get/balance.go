package get

import (
	"bytes"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jfmt"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/server/db/client"
	"github.com/memocash/server/db/item"
	"github.com/memocash/server/ref/bitcoin/tx/hs"
	"github.com/memocash/server/ref/bitcoin/tx/script"
	"github.com/memocash/server/ref/bitcoin/wallet"
)

type Balance struct {
	LockScript []byte
	Balance    int64
	Spendable  int64
	UtxoCount  int
	Spends     int
	SkipCache  bool
	OutUtxos   bool

	Utxos []*item.LockUtxo
}

func (b *Balance) GetUtxos() error {
	lockHash := script.GetLockHash(b.LockScript)
	if !b.SkipCache {
		balance, err := item.GetLockBalance(lockHash)
		if err != nil && !client.IsEntryNotFoundError(err) {
			return jerr.Get("error getting lock balance", err)
		}
		if balance != nil && !balance.NeedsSpends() {
			b.Balance = balance.Balance
			b.UtxoCount = balance.UtxoCount
			b.Spendable = balance.Spendable
			b.Spends = balance.Spends
			return nil
		}
	}
	if err := b.attachUtxos(); err != nil {
		return jerr.Get("error attaching utxos", err)
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
			b.Utxos = append(b.Utxos, lockUtxo)
		}
		if len(lockUtxos) < client.DefaultLimit {
			break
		}
		lastUid = lockUtxos[len(lockUtxos)-1].GetUid()
	}
	return nil
}

func (b *Balance) CalculateWithUtxos() error {
	var utxoValue int64
	var utxoSpendableValue int64
	var utxoSpendableCount int
	for _, utxo := range b.Utxos {
		utxoValue += utxo.Value
		if !utxo.Special {
			utxoSpendableValue += utxo.Value
			utxoSpendableCount++
		}
		if b.OutUtxos {
			jlog.Logf("UTXO: %s:%d - %s\n", hs.GetTxString(utxo.Hash), utxo.Index, utxo.Value)
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
	//jlog.Logf("address: %s\n", addr.GetEncoded())
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
	return NewBalance(lockScript), nil
}

func NewBalance(lockScript []byte) *Balance {
	return &Balance{
		LockScript: lockScript,
	}
}
