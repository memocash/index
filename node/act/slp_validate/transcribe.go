package slp_validate

import (
	"fmt"
	"github.com/jchavannes/btcd/txscript"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/node/obj/op_return/save"
	"github.com/memocash/index/ref/bitcoin/tx/parse"
	"github.com/memocash/index/ref/bitcoin/tx/slp"
)

// CascadeTranscribe is the transcription entry the cascade uses for its
// spenders. It is a package variable only so the e2e suite can instrument it
// and pin the once-per-cascade transcription guarantee (per-round
// re-transcription of fan-in spenders is the quadratic write churn that
// stalled the 2026-08-11 production audit); production never replaces it.
var CascadeTranscribe = TranscribeTxs

// TranscribeTxs runs the lenient SLP transcription for the vout-0 message of
// each tx. Validation treats a decided parent's missing output rows as
// definitive (contributing nothing), so every path that writes verdicts must
// transcribe first — including the cascade, whose reconstructed spenders may
// not have been visited by the live save path or the sweep yet.
// Objects are collected across all txs and written with a single db.Save
// (which fans out per shard/topic internally), since per-tx saves cost 2-3
// round trips each and dominate bulk backfill time.
func TranscribeTxs(slpTxs []Tx) error {
	var objects []db.Object
	for _, tx := range slpTxs {
		// Only the vout-0 message is an SLP action; later-output lokads are
		// not transcribed (spec Consideration A)
		if len(tx.Outputs) == 0 || !slp.HasSlpLokad(tx.Outputs[0].PkScript) {
			continue
		}
		pushData, err := txscript.PushedData(tx.Outputs[0].PkScript)
		if err != nil {
			continue
		}
		txObjects, err := save.TranscribeSlpObjects(parse.OpReturn{
			Saved:       true,
			TxHash:      tx.TxHash,
			PushData:    pushData,
			Outputs:     tx.Outputs,
			Inputs:      tx.Inputs,
			OutputIndex: 0,
		})
		if err != nil {
			return fmt.Errorf("error transcribing slp tx; %w", err)
		}
		objects = append(objects, txObjects...)
	}
	if len(objects) == 0 {
		return nil
	}
	if err := db.Save(objects); err != nil {
		return fmt.Errorf("error saving slp transcription objects; %w", err)
	}
	return nil
}
