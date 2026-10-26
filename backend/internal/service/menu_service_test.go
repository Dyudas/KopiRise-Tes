package service

import (
	"testing"
)

func TestGenerateSlug(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"Kopi Susu Aren Semeru", "kopi-susu-aren-semeru"},
		{"V60 Single Origin (Ijen Raung) - Special!", "v60-single-origin-ijen-raung-special"},
		{"  Matcha Latte @ Uji Kyoto  ", "matcha-latte-uji-kyoto"},
		{"French Fries & Truffle 100%", "french-fries-truffle-100"},
	}

	for _, c := range cases {
		got := generateSlug(c.input)
		if got != c.expected {
			t.Errorf("generateSlug(%q) = %q; ingin %q", c.input, got, c.expected)
		}
	}
}
