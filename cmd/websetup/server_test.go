package main

import (
	"io/fs"
	"testing"
)

func TestEmbeddedWebFS(t *testing.T) {
	sub, err := fs.Sub(embeddedWebFS, "web")
	if err != nil {
		t.Fatalf("failed fs.Sub: %v", err)
	}

	filesToCheck := []string{
		"index.html",
		"style.css",
		"app.js",
		"i18n.js",
		"assets/bg.jpeg",
		"assets/icons/01_yellow_FFFF00.png",
	}

	for _, f := range filesToCheck {
		if _, err := fs.Stat(sub, f); err != nil {
			t.Errorf("expected embedded file %s, but stat failed: %v", f, err)
		}
	}
}
