package trayanim

import (
	"testing"
	"time"
)

func TestTrayAnimatorInitAndFrames(t *testing.T) {
	anim := NewAnimator(9999)
	err := anim.Init(0, "Test Discord")
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	scubaCount := anim.FrameCount("scuba")
	if scubaCount != 30 {
		t.Errorf("Expected 30 scuba frames, got %d", scubaCount)
	}

	syncCount := anim.FrameCount("sync")
	if syncCount != 24 {
		t.Errorf("Expected 24 sync frames, got %d", syncCount)
	}

	recCount := anim.FrameCount("recording")
	if recCount != 16 {
		t.Errorf("Expected 16 recording frames, got %d", recCount)
	}

	if anim.DefaultIcon() == 0 {
		t.Errorf("DefaultIcon was not loaded (got 0)")
	}

	// Test animation start and stop
	anim.StartAnimation("sync", 20)
	time.Sleep(100 * time.Millisecond)

	anim.StartAnimation("scuba", 30)
	time.Sleep(100 * time.Millisecond)

	anim.StopAnimation()

	anim.SetTooltip("Idle")
	anim.RemoveTray()
}
