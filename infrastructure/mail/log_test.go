package mail

import "testing"

func TestMaskEmail(t *testing.T) {
	cases := map[string]string{
		"alice@example.com": "a***@example.com",
		"a@example.com":     "a***@example.com",
		"no-at":             "no-at",
		"":                  "",
	}
	for in, want := range cases {
		if got := maskEmail(in); got != want {
			t.Errorf("maskEmail(%q) = %q; want %q", in, got, want)
		}
	}
}
