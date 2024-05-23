package util

import (
	cryptoRand "crypto/rand"
	"fmt"
	"github.com/jchavannes/jgo/jutil"
	"log"
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
		log.Fatalf("fatal error getting secure random number; %v", err)
	}
	rand.Seed(secRand + int64(time.Now().Nanosecond()))
	_keyGenInit = true
}

func secureRandom() (int64, error) {
	key := [8]byte{}
	_, err := cryptoRand.Read(key[:])
	if err != nil {
		return 0, fmt.Errorf("error reading rand; %w", err)
	}
	return int64(jutil.GetUint64(key[:])), nil
}
