package client

import "github.com/memocash/index/db/store"

// Pattern is the LIKE-% style matcher; it lives in the store package (the
// scan engine) and is aliased here so RPC callers build filters without
// importing the store directly.
type Pattern = store.Pattern

var (
	// NewPatternPrefix matches uids/values starting with b (X%).
	NewPatternPrefix = store.NewPatternPrefix
	// NewPatternContains matches uids/values containing b (%X%).
	NewPatternContains = store.NewPatternContains
	// NewPatternSuffix matches uids/values ending with b (%X).
	NewPatternSuffix = store.NewPatternSuffix
)
