package op_return

import (
	"bytes"
	"context"
	"fmt"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/parse"
)

type Handler struct {
	prefix       []byte
	prefixScript []byte
	handle       func(context.Context, parse.OpReturn) error
}

func (h *Handler) CanHandle(pkScript []byte) bool {
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
		memoLikeHandler,
		memoReplyHandler,
		memoRoomPostHandler,
		memoRoomFollowHandler,
		memoRoomUnfollowHandler,
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
