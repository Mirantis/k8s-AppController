package main

import (
	"testing"
)

func TestIssueGet_Success(t *testing.T) {
	err := connect()
	if err != nil {
		t.Errorf("Error: %s", err)
	}
}
