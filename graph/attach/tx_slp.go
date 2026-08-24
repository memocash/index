package attach

import (
	"encoding/hex"
	"fmt"

	"github.com/memocash/index/db/item/slp"
	"github.com/memocash/index/graph/model"
)

// AttachSlps sets each tx's tx-level SLP action and verdict. Validity is a
// per-tx fact keyed by tx hash, so it is attached from the action and
// validity rows directly — never via token output rows, which do not exist
// for zero-quantity actions (a valid burn/no-op SEND has no SlpOutput to
// hang a verdict on). A validity row without an action row (unparseable
// vout-0 lokad message) still surfaces, with a nil type.
func (t *Tx) AttachSlps() {
	defer t.Wait.Done()
	if !t.HasField([]string{"slp"}) {
		return
	}
	txHashes := t.GetTxHashes(false, false)
	if len(txHashes) == 0 {
		return
	}
	geneses, err := slp.GetGeneses(t.Ctx, txHashes)
	if err != nil {
		t.AddError(fmt.Errorf("error getting slp geneses for attach to txs; %w", err))
		return
	}
	mints, err := slp.GetMints(t.Ctx, txHashes)
	if err != nil {
		t.AddError(fmt.Errorf("error getting slp mints for attach to txs; %w", err))
		return
	}
	sends, err := slp.GetSends(t.Ctx, txHashes)
	if err != nil {
		t.AddError(fmt.Errorf("error getting slp sends for attach to txs; %w", err))
		return
	}
	validities, err := slp.GetValidities(t.Ctx, txHashes)
	if err != nil {
		t.AddError(fmt.Errorf("error getting slp validities for attach to txs; %w", err))
		return
	}
	type action struct {
		actionType model.SlpActionType
		tokenHash  [32]byte
	}
	var actions = make(map[[32]byte]action)
	for _, genesis := range geneses {
		actions[genesis.TxHash] = action{model.SlpActionTypeGenesis, genesis.TxHash}
	}
	for _, mint := range mints {
		actions[mint.TxHash] = action{model.SlpActionTypeMint, mint.TokenHash}
	}
	for _, send := range sends {
		actions[send.TxHash] = action{model.SlpActionTypeSend, send.TokenHash}
	}
	var validityMap = make(map[[32]byte]*slp.Validity, len(validities))
	for _, validity := range validities {
		validityMap[validity.TxHash] = validity
	}
	t.Mutex.Lock()
	for i := range t.Txs {
		act, hasAction := actions[[32]byte(t.Txs[i].Hash)]
		validity, hasValidity := validityMap[[32]byte(t.Txs[i].Hash)]
		if !hasAction && !hasValidity {
			continue
		}
		var txSlp = &model.TxSlp{Validity: model.SlpValidityPending}
		if hasAction {
			actionType, tokenHash := act.actionType, model.Hash(act.tokenHash)
			txSlp.Type = &actionType
			txSlp.TokenHash = &tokenHash
		}
		if hasValidity {
			if validity.IsValid() {
				txSlp.Validity = model.SlpValidityValid
			} else {
				txSlp.Validity = model.SlpValidityInvalid
			}
		}
		t.Txs[i].Slp = txSlp
	}
	t.Mutex.Unlock()
	if err := t.attachSlpGeneses(); err != nil {
		t.AddError(err)
	}
}

// attachSlpGeneses resolves each TxSlp's token genesis, mirroring
// SlpOutputs.AttachGeneses: one batched lookup by token hash, only when the
// genesis field is requested. A genesis-action tx resolves to itself; a
// TxSlp with no token hash (verdict without action) stays nil, as does a
// token whose genesis is not in the index.
func (t *Tx) attachSlpGeneses() error {
	if !t.HasField([]string{"slp.genesis"}) {
		return nil
	}
	t.Mutex.Lock()
	var tokenHashes [][32]byte
	var slpsByToken = make(map[[32]byte][]*model.TxSlp)
	for i := range t.Txs {
		txSlp := t.Txs[i].Slp
		if txSlp == nil || txSlp.TokenHash == nil {
			continue
		}
		tokenHash := [32]byte(*txSlp.TokenHash)
		if len(slpsByToken[tokenHash]) == 0 {
			tokenHashes = append(tokenHashes, tokenHash)
		}
		slpsByToken[tokenHash] = append(slpsByToken[tokenHash], txSlp)
	}
	t.Mutex.Unlock()
	if len(tokenHashes) == 0 {
		return nil
	}
	geneses, err := slp.GetGeneses(t.Ctx, tokenHashes)
	if err != nil {
		return fmt.Errorf("error getting slp geneses for attach to tx slps; %w", err)
	}
	var allGeneses []*model.SlpGenesis
	t.Mutex.Lock()
	for _, genesis := range geneses {
		for _, txSlp := range slpsByToken[genesis.TxHash] {
			txSlp.Genesis = &model.SlpGenesis{
				Hash:       genesis.TxHash,
				TokenType:  model.Uint8(genesis.TokenType),
				Decimals:   model.Uint8(genesis.Decimals),
				BatonIndex: genesis.BatonIndex,
				Ticker:     genesis.Ticker,
				Name:       genesis.Name,
				DocURL:     genesis.DocUrl,
				DocHash:    hex.EncodeToString(genesis.DocHash[:]),
			}
			allGeneses = append(allGeneses, txSlp.Genesis)
		}
	}
	t.Mutex.Unlock()
	if err := ToSlpGeneses(t.Ctx, GetPrefixFields(t.Fields, "slp.genesis."), allGeneses); err != nil {
		return fmt.Errorf("error attaching to slp geneses for tx slps; %w", err)
	}
	return nil
}
