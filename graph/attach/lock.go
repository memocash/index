package attach

import (
	"context"
	"fmt"
	"time"

	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/item/addr"
	"github.com/memocash/index/graph/model"
	"github.com/memocash/index/ref/bitcoin/memo"
)

type Lock struct {
	base
	Locks []*model.Lock
}

func ToLocks(ctx context.Context, fields []Field, locks []*model.Lock) error {
	t := Lock{
		base:  base{Ctx: ctx, Fields: fields},
		Locks: locks,
	}
	t.Wait.Add(2)
	go t.AttachProfiles()
	go t.AttachTxs()
	t.Wait.Wait()
	if len(t.Errors) > 0 {
		return fmt.Errorf("error attaching details to txs; %w", t.Errors[0])
	}
	return nil
}

func (l *Lock) GetLockAddrs() [][25]byte {
	l.Mutex.Lock()
	defer l.Mutex.Unlock()
	var lockAddrs [][25]byte
	for _, lock := range l.Locks {
		lockAddrs = append(lockAddrs, lock.Address)
	}
	return lockAddrs
}

func (l *Lock) AttachProfiles() {
	defer l.Wait.Done()
	if !l.HasField([]string{"profile"}) {
		return
	}
	var allProfiles []*model.Profile
	l.Mutex.Lock()
	for _, lock := range l.Locks {
		lock.Profile = &model.Profile{Address: lock.Address}
		allProfiles = append(allProfiles, lock.Profile)
	}
	l.Mutex.Unlock()
	if err := ToMemoProfiles(l.Ctx, GetPrefixFields(l.Fields, "profile"), allProfiles); err != nil {
		l.AddError(fmt.Errorf("error attaching to lock profiles; %w", err))
		return
	}
}

func (l *Lock) AttachTxs() {
	defer l.Wait.Done()
	if !l.HasField([]string{"txs"}) {
		return
	}
	l.Mutex.Lock()
	defer l.Mutex.Unlock()
	txsField := l.Fields.GetField("txs")
	startDate, _ := model.UnmarshalDate(txsField.Arguments["start"])
	startTx, _ := txsField.Arguments["tx"].(string)
	limit, _ := model.UnmarshalUint32(txsField.Arguments["limit"])
	var allTxs []*model.Tx
	for _, lock := range l.Locks {
		var startUid []byte
		if time.Time(startDate).After(memo.GetGenesisTime()) {
			startUid = jutil.CombineBytes(lock.Address[:], jutil.GetTimeByteNanoBig(time.Time(startDate)))
			if len(startTx) > 0 {
				txHash, err := chainhash.NewHashFromStr(startTx)
				if err != nil {
					l.AddError(fmt.Errorf("error decoding start hash for lock txs resolver; %w", err))
					return
				}
				startUid = append(startUid, jutil.ByteReverse(txHash[:])...)
			}
		}
		seenTxs, err := addr.GetSeenTxs(l.Ctx, lock.Address, startUid, uint32(limit))
		if err != nil {
			l.AddError(fmt.Errorf("error getting addr seen txs for lock txs resolver; %w", err))
			return
		}
		lock.Txs = make([]*model.Tx, len(seenTxs))
		for i := range seenTxs {
			lock.Txs[i] = &model.Tx{
				Hash: seenTxs[i].TxHash,
				Seen: model.Date(seenTxs[i].Seen),
			}
		}
		allTxs = append(allTxs, lock.Txs...)
	}
	if err := ToTxs(l.Ctx, txsField.Fields, allTxs); err != nil {
		l.AddError(fmt.Errorf("error attaching to lock transactions; %w", err))
		return
	}
}
