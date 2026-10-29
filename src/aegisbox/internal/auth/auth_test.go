package auth

import "testing"

func TestCheckPasswordHash(t *testing.T) {
	hashed, err := HashPassword("pass1")
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name, password, stored string
		want                   bool
	}{
		{"bcrypt match", "pass1", hashed, true},
		{"bcrypt mismatch", "pass2", hashed, false},
		{"legacy plain match", "pass1", "pass1", true},
		{"legacy plain mismatch", "pass2", "pass1", false},
		{"plain password equal to the hash text", hashed, hashed, false},
	}
	for _, c := range cases {
		if got := CheckPasswordHash(c.password, c.stored); got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
	if !IsBcryptHash(hashed) || IsBcryptHash("pass1") {
		t.Error("IsBcryptHash misclassified a value")
	}
}
