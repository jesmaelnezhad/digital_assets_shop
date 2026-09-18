package handlers

import "testing"

func TestExtractHTTPURL(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"hello", ""},
		{"see https://example.com/pack and more", "https://example.com/pack"},
		{"https://example.com/a.", "https://example.com/a"},
		{"ftp://example.com/x", ""},
		{"(https://store.example/x)", "https://store.example/x"},
	}
	for _, c := range cases {
		if got := extractHTTPURL(c.in); got != c.want {
			t.Fatalf("extractHTTPURL(%q)=%q want %q", c.in, got, c.want)
		}
	}
}

func TestHostBlocked(t *testing.T) {
	blocked := []string{"localhost", "127.0.0.1", "10.1.2.3", "192.168.0.9", "169.254.169.254", "::1", "metadata.google.internal"}
	for _, h := range blocked {
		if !hostBlocked(h) {
			t.Fatalf("expected blocked %q", h)
		}
	}
	if hostBlocked("example.com") {
		t.Fatal("example.com should be allowed before DNS")
	}
	if publicLinkURL("http://127.0.0.1/secret") != "" {
		t.Fatal("loopback URL must not become a post card")
	}
	if publicLinkURL("https://example.com/kit") != "https://example.com/kit" {
		t.Fatal("public URL should stay")
	}
}

func TestSanitizeImageURL(t *testing.T) {
	ok := "data:image/png;base64,iVBORw0KGgo="
	if sanitizeImageURL(ok) != ok {
		t.Fatal("png data URL")
	}
	if sanitizeImageURL("data:image/svg+xml;base64,PHN2Zz4=") != "" {
		t.Fatal("svg data URL must be rejected")
	}
	if sanitizeImageURL("javascript:alert(1)") != "" {
		t.Fatal("javascript URL")
	}
	if sanitizeImageURL("https://cdn.example/p.jpg") != "https://cdn.example/p.jpg" {
		t.Fatal("https image")
	}
	if sanitizeImageURL("http://127.0.0.1/x.png") != "" {
		t.Fatal("loopback image")
	}
}

func TestParseLinkMeta(t *testing.T) {
	html := `<html><head>
<title>Page Title</title>
<meta property="og:title" content="OG Title" />
<meta property="og:description" content="A pack of clay." />
<meta property="og:image" content="/hero.jpg" />
</head></html>`
	title, desc, image := parseLinkMeta(html, "https://example.com/post")
	if title != "OG Title" {
		t.Fatalf("title=%q", title)
	}
	if desc != "A pack of clay." {
		t.Fatalf("desc=%q", desc)
	}
	if image != "https://example.com/hero.jpg" {
		t.Fatalf("image=%q", image)
	}

	fallback := `<html><head><title>Just Title</title></head></html>`
	title, desc, image = parseLinkMeta(fallback, "https://example.com/")
	if title != "Just Title" || desc != "" || image != "" {
		t.Fatalf("fallback title=%q desc=%q image=%q", title, desc, image)
	}
}
