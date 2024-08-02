package client

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/jchavannes/jgo/db_util"
	"github.com/jchavannes/jgo/jutil"
	"time"
)

const (
	DefaultGetTimeout    = 60 * time.Second
	DefaultSetTimeout    = 10 * time.Minute
	DefaultStreamTimeout = 7 * 24 * time.Hour
)

const (
	DefaultLimit = db_util.DefaultLimit
	MediumLimit  = db_util.MediumLimit
	LargeLimit   = db_util.LargeLimit
	ExLargeLimit = db_util.ExLargeLimit
	HugeLimit    = db_util.HugeLimit
)

const (
	MaxMessageSize = 32 * 10e7
)

const (
	MessageNotSetErrorMessage  = "error message not set"
	MultipleEntryErrorMessage  = "error multiple entries found"
	EntryNotFoundErrorMessage  = "error entry not found"
	ResourceUnavailableMessage = "resource temporarily unavailable"
)

var (
	MultipleEntryError       = fmt.Errorf(MultipleEntryErrorMessage)
	EntryNotFoundError       = fmt.Errorf(EntryNotFoundErrorMessage)
	MessageNotSetError       = fmt.Errorf(MessageNotSetErrorMessage)
	ResourceUnavailableError = fmt.Errorf(ResourceUnavailableMessage)
)

type Topic struct {
	Name string
	Size uint64
}

type Message struct {
	Topic   string
	Uid     []byte
	Message []byte
}

func (m Message) UidHex() string {
	return hex.EncodeToString(m.Uid)
}

func (m Message) MessageHex() string {
	return hex.EncodeToString(m.Message)
}

func GenShardSource(b []byte) uint {
	return jutil.GetByteMd5Int(b)
}

func GenShardSource32(b []byte) uint32 {
	return uint32(GenShardSource(b))
}

func IncrementBytes(b []byte) []byte {
	var r []byte
	for i := len(b) - 1; i >= 0; i-- {
		c := b[i]
		if c < 0xff {
			r = make([]byte, i+1)
			copy(r, b)
			r[i] = c + 1
			break
		}
	}
	return r
}

func IsMultipleEntryError(e error) bool {
	return errors.Is(e, MultipleEntryError)
}

func IsEntryNotFoundError(e error) bool {
	return errors.Is(e, EntryNotFoundError)
}

func IsMessageNotSetError(err error) bool {
	return errors.Is(err, MessageNotSetError)
}

func IsResourceUnavailableError(err error) bool {
	return errors.Is(err, ResourceUnavailableError)
}
