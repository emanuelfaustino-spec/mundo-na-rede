//go:build windows

package main

import "testing"

func TestNativeTitlebarColorOrder(t *testing.T) {
	tests := []struct {
		name string
		r, g, b byte
		want uint32
	}{
		{"dark_caption", 9, 13, 29, 0x001D0D09},
		{"light_caption", 245, 247, 252, 0x00FCF7F5},
		{"white", 255, 255, 255, 0x00FFFFFF},
	}
	for _, test := range tests {
		if got := nativeRGB(test.r, test.g, test.b); got != test.want {
			t.Errorf("%s: COLORREF = %#08x; want %#08x", test.name, got, test.want)
		}
	}
}
