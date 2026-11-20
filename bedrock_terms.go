package main

import "net/http"

// Both the original per-server setup and the shared engine library can
// provide explicit acceptance of the Minecraft Bedrock terms.
func acceptedBDSTerms(r *http.Request) bool {
	return r.Header.Get("X-Accept-BDS-Terms") == "yes" || r.Header.Get("X-Accept-Minecraft-Terms") == "yes"
}
