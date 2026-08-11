package op_return

import (
	"context"
	"fmt"
	"github.com/memocash/index/node/act/slp_validate"
	"github.com/memocash/index/node/obj/op_return/save"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/parse"
	"github.com/memocash/index/ref/bitcoin/tx/slp"
)

var slpTokenHandler = &Handler{
	prefix:    memo.PrefixSlp,
	canHandle: slp.HasSlpLokad,
	handle: func(ctx context.Context, info parse.OpReturn) error {
		// An SLP action is defined entirely by the vout-0 message; lokad
		// bytes in any later output are not an SLP action (spec Consideration
		// A) and must neither be transcribed nor validated. Ignoring them
		// (rather than transcribing or marking invalid) keeps a tx's SLP
		// standing a function of its own vout 0 alone: a later-output lokad
		// can never write token rows, and can never flip its verdict based on
		// whether vout 0 happens to be SLP.
		if info.OutputIndex != 0 {
			return nil
		}
		if err := save.TranscribeSlp(info); err != nil {
			return err
		}
		// Strict parsing works from the raw vout-0 script, independent of the
		// lenient transcription above.
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
