package git

import (
	"os"
	"path/filepath"
	"testing"
)

func TestClone(t *testing.T) {
	t.Skip("Skipping TestClone")

	base := os.TempDir()
	os.MkdirAll(base, 0666)
	dir := filepath.Join(base, "goboot-starter")
	url := "https://github.com/gostones/goboot-starter.git"
	Clone(url, dir)

	t.Logf("dir: %v", dir)
}
