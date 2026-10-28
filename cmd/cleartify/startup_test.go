package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestListenerFailureExitsNonzero(t *testing.T) {
	if os.Getenv("CLEARTIFY_TEST_STARTUP") == "1" {
		main()
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestListenerFailureExitsNonzero$")
	// Supply no provider configuration: this is a disposable local process.
	cmd.Env = []string{"PATH=" + os.Getenv("PATH"), "HOME=" + os.Getenv("HOME"), "TURSO_GO_CACHE_DIR=" + os.Getenv("TURSO_GO_CACHE_DIR"), "CLEARTIFY_TEST_STARTUP=1", "PORT=invalid-port", "DATABASE_PATH=" + t.TempDir() + "/db", "APP_SECRET=" + strings.Repeat("x", 32)}
	output, e := cmd.CombinedOutput()
	if e == nil {
		t.Fatalf("failed listener exited successfully: %s", output)
	}
	if !strings.Contains(string(output), "http server failed") {
		t.Fatalf("process failed before listener: %s", output)
	}
}
