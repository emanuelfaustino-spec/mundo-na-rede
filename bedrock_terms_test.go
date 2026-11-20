package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBedrockTermsAcceptedByBothInstallationFlows(t *testing.T) {
	cases := []struct {
		name, header, value string
		want                bool
	}{
		{"library consent", "X-Accept-Minecraft-Terms", "yes", true},
		{"older server flow", "X-Accept-BDS-Terms", "yes", true},
		{"missing consent", "", "", false},
		{"rejected consent", "X-Accept-Minecraft-Terms", "no", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/engines/prepare?kind=bedrock", nil)
			if tc.header != "" {
				req.Header.Set(tc.header, tc.value)
			}
			if actual := acceptedBDSTerms(req); actual != tc.want {
				t.Fatalf("acceptedBDSTerms=%v, expected %v", actual, tc.want)
			}
		})
	}
}
