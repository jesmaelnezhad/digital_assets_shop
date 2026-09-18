package handlers

import "testing"

func TestAllowedExtensionsCoverShopAssets(t *testing.T) {
	for _, ext := range []string{".jpg", ".jpeg", ".png", ".webp", ".gif", ".zip", ".pdf", ".svg"} {
		if _, ok := allowedExtensions[ext]; !ok {
			t.Fatalf("missing extension %s", ext)
		}
	}
	if _, ok := allowedExtensions[".exe"]; ok {
		t.Fatal("executables must not be allowed")
	}
}
