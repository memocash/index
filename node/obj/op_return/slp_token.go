package op_return

import (
	"context"
	"fmt"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/node/act/slp_validate"
	"github.com/memocash/index/node/obj/op_return/save"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/parse"
	"github.com/memocash/index/ref/bitcoin/tx/slp"
)

var slpTokenHandler = &Handler{
	prefix:    memo.PrefixSlp,
	noAddr:    true,
	canHandle: slp.HasSlpLokad,
	handle: func(ctx context.Context, info parse.OpReturn) error {
		if err := TranscribeSlp(info); err != nil {
			return err
		}
		// Validity is a property of the vout-0 message; when the handled
		// output isn't vout 0 and vout 0 is itself SLP, that invocation
		// decides. Otherwise validate here (strict parsing works from the
		// raw script, independent of the lenient transcription above).
		if info.OutputIndex != 0 && len(info.Outputs) > 0 && slp.HasSlpLokad(info.Outputs[0].PkScript) {
			return nil
		}
		if _, err := slp_validate.ValidateTxsCascade(ctx, []slp_validate.Tx{{
			TxHash:  info.TxHash,
			Inputs:  info.Inputs,
			Outputs: info.Outputs,
		}}); err != nil {
			return fmt.Errorf("error validating slp tx; %w", err)
		}
		return nil
	},
}

// TranscribeSlp runs the lenient SLP transcription for one SLP op_return
// output. Exposed for the validity sweeper, which re-runs transcription for
// txs the live path may have missed before validating them.
func TranscribeSlp(info parse.OpReturn) error {
	if len(info.PushData) < 5 {
		if err := item.LogProcessError(&item.ProcessError{
			TxHash: info.TxHash,
			Error:  fmt.Sprintf("invalid slp, incorrect push data (%d) op return handler", len(info.PushData)),
		}); err != nil {
			return fmt.Errorf("error saving process error for slp incorrect push data; %w", err)
		}
		return nil
	}
	switch memo.SlpType(info.PushData[2]) {
	case memo.SlpTxTypeGenesis:
		if err := save.SlpGenesis(info); err != nil {
			return fmt.Errorf("error saving slp genesis op return handler; %w", err)
		}
	case memo.SlpTxTypeMint:
		if err := save.SlpMint(info); err != nil {
			return fmt.Errorf("error saving slp mint op return handler; %w", err)
		}
	case memo.SlpTxTypeSend:
		if err := save.SlpSend(info); err != nil {
			return fmt.Errorf("error saving slp send op return handler; %w", err)
		}
	case memo.SlpTxTypeCommit:
		if err := save.SlpCommit(info); err != nil {
			return fmt.Errorf("error saving slp commit op return handler; %w", err)
		}
	default:
		if err := item.LogProcessError(&item.ProcessError{
			TxHash: info.TxHash,
			Error:  fmt.Sprintf("unknown slp tx type op return handler: %s", info.PushData[2]),
		}); err != nil {
			return fmt.Errorf("error saving process error for slp unknown tx type; %w", err)
		}
		return nil
	}
	return nil
}
