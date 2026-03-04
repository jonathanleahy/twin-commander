package main

import "time"

// KeyAction represents a completed multi-key command.
type KeyAction int

const (
	KeyActionNone     KeyAction = iota
	KeyActionJumpTop            // gg
	KeyActionDelete             // dd
	KeyActionYank               // yy
	KeyActionGitStage           // gs
	KeyActionGoDir              // gd
	KeyActionGoRecent           // gr
)

// KeySequence tracks multi-key sequences (vim-style gg, dd, yy, gs).
type KeySequence struct {
	pending rune
	timer   *time.Timer
}

// NewKeySequence creates a new multi-key sequence tracker.
func NewKeySequence() *KeySequence {
	return &KeySequence{}
}

// Feed processes a rune and returns (action, consumed).
// If the rune starts or completes a sequence, consumed is true.
// If a sequence is completed, action is non-zero.
func (ks *KeySequence) Feed(r rune) (KeyAction, bool) {
	if ks.timer != nil {
		ks.timer.Stop()
		ks.timer = nil
	}

	if ks.pending != 0 {
		first := ks.pending
		ks.pending = 0

		switch {
		case first == 'g' && r == 'g':
			return KeyActionJumpTop, true
		case first == 'g' && r == 's':
			return KeyActionGitStage, true
		case first == 'g' && r == 'd':
			return KeyActionGoDir, true
		case first == 'g' && r == 'r':
			return KeyActionGoRecent, true
		case first == 'd' && r == 'd':
			return KeyActionDelete, true
		case first == 'y' && r == 'y':
			return KeyActionYank, true
		}
		// Sequence didn't match — drop both keys
		return KeyActionNone, false
	}

	// Start of a potential sequence
	if r == 'g' || r == 'd' || r == 'y' {
		ks.pending = r
		ks.timer = time.AfterFunc(500*time.Millisecond, func() {
			ks.pending = 0
		})
		return KeyActionNone, true
	}

	return KeyActionNone, false
}
