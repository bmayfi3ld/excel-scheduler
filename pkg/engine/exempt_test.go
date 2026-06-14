package engine

import "testing"

func TestIsExempt(t *testing.T) {
	cases := []struct {
		input string
		want  bool
	}{
		{"####", true},
		{"#### blocked", true},
		{"@@@@ no class", true},
		{"**** closed", true},
		{"###x", false},
		{"###", false}, // only 3 chars
		{"", false},
		{"1st", false},
		{"Latin Cart", false},
		{"@@@@", true},
		{"aaaa", true},
		{"aaab", false},
		{"aaa", false},
	}
	for _, c := range cases {
		got := isExempt(c.input)
		if got != c.want {
			t.Errorf("isExempt(%q) = %v, want %v", c.input, got, c.want)
		}
	}
}
