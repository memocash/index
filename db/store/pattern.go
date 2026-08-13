package store

import "bytes"

// Pattern parts are matched in order with arbitrary gaps between them (LIKE with %):
//
//	prefix  X%   => Parts:[X], AnchorStart
//	contains %X% => Parts:[X]
//	suffix  %X   => Parts:[X], AnchorEnd
//	A%B          => Parts:[A,B], AnchorStart, AnchorEnd
type Pattern struct {
	Parts       [][]byte
	AnchorStart bool
	AnchorEnd   bool
}

// Match reports whether b matches the pattern. A nil or part-less pattern matches everything.
func (p *Pattern) Match(b []byte) bool {
	if p == nil || len(p.Parts) == 0 {
		return true
	}
	parts := p.Parts
	if p.AnchorStart {
		if !bytes.HasPrefix(b, parts[0]) {
			return false
		}
		if len(parts) == 1 {
			if p.AnchorEnd {
				return len(b) == len(parts[0])
			}
			return true
		}
		return matchParts(b[len(parts[0]):], parts[1:], p.AnchorEnd)
	}
	return matchParts(b, parts, p.AnchorEnd)
}

// matchParts matches parts left to right with arbitrary gaps; the last part must be a
// non-overlapping suffix when anchorEnd is set.
func matchParts(b []byte, parts [][]byte, anchorEnd bool) bool {
	var last = len(parts)
	if anchorEnd {
		last--
	}
	var pos int
	for i := 0; i < last; i++ {
		idx := bytes.Index(b[pos:], parts[i])
		if idx < 0 {
			return false
		}
		pos += idx + len(parts[i])
	}
	if anchorEnd {
		lastPart := parts[len(parts)-1]
		return len(b)-pos >= len(lastPart) && bytes.HasSuffix(b, lastPart)
	}
	return true
}

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
