package main

import (
	"testing"
	"path/filepath"
)

func TestFilterIpV4(t *testing.T){
}

func TestFolderOfPanic1(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	url := "http://foo.bar/.."
	FolderOf(url)
}

func TestFolderOfPanic2(t *testing.T) {
	url := "http://foo.bar/../../../foobar"
	u := FolderOf(url)
	if filepath.Base(u) != "foobar" {
		t.Fatalf("url of return incorrect value")
	}
}

func TestFolderOfNormal(t *testing.T) {
	url := "http://foo.bar/file"
	u := FolderOf(url)
	if filepath.Base(u) != "file" {
		t.Fatalf("url of return incorrect value")
	}
}
