package core

import "testing"

func TestIsValidUserSecret(t *testing.T) {
	cases := []struct {
		name   string
		secret string
		want   bool
	}{
		{name: "valid", secret: "Admin123!", want: true},
		{name: "too short", secret: "A1!", want: false},
		{name: "missing digit", secret: "AdminOnly!", want: false},
		{name: "missing special", secret: "Admin123", want: false},
		{name: "empty", secret: "", want: false},
	}
	for _, tc := range cases {
		if got := IsValidUserSecret(tc.secret); got != tc.want {
			t.Fatalf("%s: IsValidUserSecret()=%v want %v", tc.name, got, tc.want)
		}
	}
}
