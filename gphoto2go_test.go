package gphoto2go

import "testing"

func TestCapturePreviewSanity(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatal("Expected capture to work")
		}
	}()
	cam := &Camera{}
	cam.Init()
	_, i := cam.CapturePreview()
	if i != 0 {
		t.Fatalf("Expected 0, got %d. Camera must be on", i)
	}
}
