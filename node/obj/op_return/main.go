package op_return

import (
	"bytes"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/ref/bitcoin/memo"
)

type Info struct {
	Height   int64
	TxHash   []byte
	LockHash []byte
	PushData [][]byte
}

type Handler struct {
	prefix       []byte
	prefixScript []byte
	handle       func(Info) error
}

func (h *Handler) CanHandle(pkScript []byte) bool {
	return len(pkScript) >= len(h.prefixScript) &&
		bytes.Equal(pkScript[:len(h.prefixScript)], h.prefixScript)
}

func (h *Handler) Handle(info Info) error {
	if h.handle == nil {
		return jerr.Newf("error handler not set (prefix: %x)", h.prefix)
	}
	if err := h.handle(info); err != nil {
		return jerr.Getf(err, "error processing op return handler (prefix: %x)", h.prefix)
	}
	return nil
}

func GetHandlers() ([]*Handler, error) {
	var handlers = []*Handler{
		memoNameHandler,
		memoProfileHandler,
		memoProfilePicHandler,
		memoFollowHandler,
	}
	for _, opReturn := range handlers {
		prefixScript, err := memo.GetBaseOpReturn().AddData(opReturn.prefix).Script()
		if err != nil {
			return nil, jerr.Get("error getting script for memo code", err)
		}
		opReturn.prefixScript = prefixScript
	}
	return handlers, nil
}
