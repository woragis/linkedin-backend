package password

import "testing"

func TestHashAndCompare(t *testing.T) {
	plain := "demo12345"
	hash, err := Hash(plain)
	if err != nil {
		t.Fatal(err)
	}
	if hash == plain {
		t.Fatal("hash should not equal plaintext")
	}
	if !Compare(hash, plain) {
		t.Fatal("Compare should succeed for correct password")
	}
	if Compare(hash, "wrong-password") {
		t.Fatal("Compare should fail for wrong password")
	}
}

func TestCompare_EmptyHash(t *testing.T) {
	if Compare("", "anything") {
		t.Fatal("empty hash should not match")
	}
}
