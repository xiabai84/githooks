package hooks

import (
	"io"
	"os"
	"testing"
)

// captureStdout runs fn while capturing stdout and returns the captured output.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe failed: %v", err)
	}
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = oldStdout

	captured, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("io.ReadAll failed: %v", err)
	}
	return string(captured)
}
