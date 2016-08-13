package main

import (
	"testing"
	"os/user"
	"path/filepath"
)

func TestPartCalculate(t *testing.T) {
	displayProgress = false

	parts := partCalculate(int64(10), 100, "http://foo.bar/file")
	if len(parts) != 10 {
		t.Fatalf("parts length should be 10")
	}
	if parts[0].Url != "http://foo.bar/file" {
		t.Fatalf("part url was wrong")
	}
	usr, _ := user.Current()
	dir := filepath.Join(usr.HomeDir, dataFolder, "file/file.part0")
	if parts[0].Path != dir {
		t.Fatalf("part path was wrong")
	}
	if parts[0].RangeFrom != 0 && parts[0].RangeTo != 10 {
		t.Fatalf("part range was wrong")
	}
}

func TestNewHttpDownloader(t *testing.T) {
	displayProgress = false

	NewHttpDownloader("http://www.golangtc.com:80/static/go/1.7rc6/go1.7rc6.darwin-amd64.pkg", 1, true)
	NewHttpDownloader("http://www.golangtc.com/static/go/1.7rc6/go1.7rc6.darwin-amd64.pkg", 1, true)
}
