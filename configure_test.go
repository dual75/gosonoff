package main

import (
	"testing"
)

func TestConfigure(t *testing.T) {
	sid, pwd := "dummysid", "dummypassword"
	err := configure("http://posttestserver.com/post.php", &sid, &pwd)
	if err != nil {
		t.Fatalf("Expected nil error, got: %v", err)
	}
	return
}
