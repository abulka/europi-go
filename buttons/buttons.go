package buttons

import "time"

// DigitalInput abstracts a physical button.
type DigitalInput interface {
	Pressed() bool
}

// Event represents a button press event.
type Event int

const (
	None Event = iota
	B1Press
	B2Press
)

type ButtonManager struct {
	b1, b2         DigitalInput
	b1Pressed      bool
	b2Pressed      bool
	b1ChangedAt    time.Time
	b2ChangedAt    time.Time
	b1Start        time.Time
	b2Start        time.Time
	debounce       time.Duration
	heldThreshold  time.Duration
}

func New(b1, b2 DigitalInput) *ButtonManager {
	return &ButtonManager{
		b1:            b1,
		b2:            b2,
		debounce:      50 * time.Millisecond,
		heldThreshold: 1 * time.Second,
	}
}

// Update reads button states and returns any press events.
func (bm *ButtonManager) Update() Event {
	now := time.Now()

	// --- B1 logic ---
	b1Now := bm.b1.Pressed()
	if b1Now != bm.b1Pressed && now.Sub(bm.b1ChangedAt) > bm.debounce {
		bm.b1ChangedAt = now
		bm.b1Pressed = b1Now
		if b1Now {
			bm.b1Start = now
		} else if !bm.BothHeld() {
			return B1Press
		}
	}

	// --- B2 logic ---
	b2Now := bm.b2.Pressed()
	if b2Now != bm.b2Pressed && now.Sub(bm.b2ChangedAt) > bm.debounce {
		bm.b2ChangedAt = now
		bm.b2Pressed = b2Now
		if b2Now {
			bm.b2Start = now
		} else if !bm.BothHeld() {
			return B2Press
		}
	}

	return None
}

// BothHeld returns true if both buttons have been held for threshold duration.
func (bm *ButtonManager) BothHeld() bool {
	if bm.b1Pressed && bm.b2Pressed {
		since := time.Since(max(bm.b1Start, bm.b2Start))
		return since >= bm.heldThreshold
	}
	return false
}

func max(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}
