package slp_validate

import (
	"fmt"
	"github.com/jchavannes/btcd/txscript"
	"github.com/memocash/index/node/obj/op_return/save"
	"github.com/memocash/index/ref/bitcoin/tx/parse"
	"github.com/memocash/index/ref/bitcoin/tx/slp"
)

// TranscribeTxs runs the lenient SLP transcription for the vout-0 message of
// each tx. Validation treats a decided parent's missing output rows as
// definitive (contributing nothing), so every path that writes verdicts must
// transcribe first — including the cascade, whose reconstructed spenders may
// not have been visited by the live save path or the sweep yet.
func TranscribeTxs(slpTxs []Tx) error {
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
		if err := save.TranscribeSlp(parse.OpReturn{
			Saved:       true,
			TxHash:      tx.TxHash,
			PushData:    pushData,
			Outputs:     tx.Outputs,
			Inputs:      tx.Inputs,
			OutputIndex: 0,
		}); err != nil {
			return fmt.Errorf("error transcribing slp tx; %w", err)
		}
	}
	return nil
}
