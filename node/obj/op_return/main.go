package op_return

import (
	"bytes"
	"context"
	"fmt"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/parse"
	"github.com/memocash/index/ref/bitcoin/wallet"
)

type Handler struct {
	prefix       []byte
	prefixScript []byte
	canHandle    func(pkScript []byte) bool
	handle       func(context.Context, parse.OpReturn) error
}

// getSenderAddr resolves the sender address for a memo op return: the first
// input whose unlock script yields an address, matching the memo.cash site
// rule. If no input yields one, a process error is logged and a nil addr is
// returned; the handler should skip the tx. This is a memo protocol
// convention, not a framework rule — SLP transcription/validation is
// address-independent and never calls this, so SLP txs are indexed even when
// no input unlock script yields an address (e.g. P2SH inputs).
func getSenderAddr(info parse.OpReturn) (*wallet.Addr, error) {
	for _, in := range info.Inputs {
		if addr, err := wallet.GetAddrFromUnlockScript(in.SignatureScript); err == nil {
			return addr, nil
		}
	}
	if err := item.LogProcessError(&item.ProcessError{
		TxHash: info.TxHash,
		Error:  fmt.Sprintf("error could not find input pk hash for op return: %s", chainhash.Hash(info.TxHash)),
	}); err != nil {
		return nil, fmt.Errorf("error saving process error for op return without input address; %w", err)
	}
	return nil, nil
}

func (h *Handler) CanHandle(pkScript []byte) bool {
	if h.canHandle != nil {
		return h.canHandle(pkScript)
	}
	return len(pkScript) >= len(h.prefixScript) &&
		bytes.Equal(pkScript[:len(h.prefixScript)], h.prefixScript)
}

func (h *Handler) Handle(ctx context.Context, info parse.OpReturn) error {
	if h.handle == nil {
		return fmt.Errorf("error handler not set (prefix: %x)", h.prefix)
	}
	if err := h.handle(ctx, info); err != nil {
		return fmt.Errorf("error processing op return handler (prefix: %x); %w", h.prefix, err)
	}
	return nil
}

func GetHandlers() ([]*Handler, error) {
	var handlers = []*Handler{
		memoNameHandler,
		memoProfileHandler,
		memoProfilePicHandler,
		memoFollowHandler,
		memoUnfollowHandler,
		memoPostHandler,
		memoSendHandler,
		memoLikeHandler,
		memoReplyHandler,
		memoRoomPostHandler,
		memoRoomFollowHandler,
		memoRoomUnfollowHandler,
		memoLinkRequestHandler,
		memoLinkAcceptHandler,
		memoLinkRevokeHandler,
		slpTokenHandler,
	}
	for _, opReturn := range handlers {
		prefixScript, err := memo.GetBaseOpReturn().AddData(opReturn.prefix).Script()
		if err != nil {
			return nil, fmt.Errorf("error getting script for memo code; %w", err)
		}
		opReturn.prefixScript = prefixScript
	}
	return handlers, nil
}
