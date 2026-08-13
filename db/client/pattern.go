package client

import "github.com/memocash/index/db/store"

// Pattern is the LIKE-% style matcher; the type lives in the store package
// (the scan engine matches against it directly) and is aliased here so RPC
// callers build filters without importing the store. The constructors are
// client-only: the store never builds patterns, it only matches them.
type Pattern = store.Pattern

// NewPatternPrefix matches uids/values starting with b (X%).
func NewPatternPrefix(b []byte) *Pattern {
	return &Pattern{Parts: [][]byte{b}, AnchorStart: true}
}

// NewPatternContains matches uids/values containing b (%X%).
func NewPatternContains(b []byte) *Pattern {
	return &Pattern{Parts: [][]byte{b}}
}

// NewPatternSuffix matches uids/values ending with b (%X).
func NewPatternSuffix(b []byte) *Pattern {
	return &Pattern{Parts: [][]byte{b}, AnchorEnd: true}
}
