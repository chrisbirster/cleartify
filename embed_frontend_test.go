package cleartify

import (
	"io/fs"
	"testing"
	"testing/fstest"
)

func TestFrontendIncludesPlaceholderIndex(t *testing.T) {
	t.Parallel()

	indexHTML, err := fs.ReadFile(Frontend(), "index.html")
	if err != nil {
		t.Fatalf("expected embedded index.html: %v", err)
	}

	if len(indexHTML) == 0 {
		t.Fatal("expected embedded index.html to be non-empty")
	}
}

func TestSubFrontendPanicsWhenDirectoryIsMissing(t *testing.T) {
	t.Parallel()

	defer func() {
		if recover() == nil {
			t.Fatal("expected subFrontend to panic for a missing directory")
		}
	}()

	subFrontend(fstest.MapFS{
		"file.txt": {Data: []byte("not a directory")},
	}, "../bad")
}
