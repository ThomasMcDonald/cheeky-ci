package main

import (
	"testing"
)

func TestGetApiKeyHash(t *testing.T) {
	token := "1a269e5d-ab18-41ef-9921-181b44a9230f"
	want := "abaf982020f45ea41e6d7f1050f1eaca"

	hash, err := GetApiKeyHash(token)

	if want != hash || err != nil {
		t.Errorf("GetApiKeyHash('%s') = %s, %v. Want %s", token, hash, err, want)
	}
}
