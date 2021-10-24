package util

import (
	cryptoRand "crypto/rand"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"math/rand"
	"time"
)

var _keyGenInit bool

func SeedRandom() {
	if _keyGenInit {
		return
	}
	secRand, err := secureRandom()
	if err != nil {
		jerr.Get("fatal error getting secure random number", err).Fatal()
	}
	rand.Seed(secRand + int64(time.Now().Nanosecond()))
	_keyGenInit = true
}

func secureRandom() (int64, error) {
	key := [8]byte{}
	_, err := cryptoRand.Read(key[:])
	if err != nil {
		return 0, jerr.Get("error reading rand", err)
	}
	return int64(jutil.GetUint64(key[:])), nil
}
