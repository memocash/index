package client

import (
	"bytes"
	"encoding/hex"
	"github.com/jchavannes/jgo/db_util"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"time"
)

const (
	DefaultGetTimeout    = 60 * time.Second
	DefaultSetTimeout    = 10 * time.Minute
	DefaultWaitTimeout   = 5 * time.Minute
	DefaultStreamTimeout = 7 * 24 * time.Hour
)

const (
	DefaultLimit = db_util.DefaultLimit
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
	MultipleEntryError       = jerr.New(MultipleEntryErrorMessage)
	EntryNotFoundError       = jerr.New(EntryNotFoundErrorMessage)
	MessageNotSetError       = jerr.New(MessageNotSetErrorMessage)
	ResourceUnavailableError = jerr.New(ResourceUnavailableMessage)
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

func GetByteShard(b []byte) uint {
	return jutil.GetByteMd5Int(b)
}

func GetByteShard32(b []byte) uint32 {
	return uint32(GetByteShard(b))
}

func GetMaxStart() []byte {
	return []byte{0xff, 0x00, 0xff, 0x00, 0xff, 0x00, 0xff, 0x00}
}

func IsMaxStart(d []byte) bool {
	return bytes.Equal(d, GetMaxStart())
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
	return jerr.HasErrorPart(e, MultipleEntryErrorMessage)
}

func IsEntryNotFoundError(e error) bool {
	return jerr.HasError(e, EntryNotFoundErrorMessage)
}

func IsMessageNotSetError(err error) bool {
	return jerr.HasError(err, MessageNotSetErrorMessage)
}

func IsResourceUnavailableError(err error) bool {
	return jerr.HasErrorPart(err, ResourceUnavailableMessage)
}
