package handlers

import (
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

var httpURLRe = regexp.MustCompile(`https?://[^\s<>"'\)\]]+`)

var (
	ogTitleRe = regexp.MustCompile(`(?is)<meta[^>]+(?:property|name)=["']og:title["'][^>]*content=["']([^"']+)["']|<meta[^>]+content=["']([^"']+)["'][^>]*(?:property|name)=["']og:title["']`)
	ogDescRe  = regexp.MustCompile(`(?is)<meta[^>]+(?:property|name)=["']og:description["'][^>]*content=["']([^"']+)["']|<meta[^>]+content=["']([^"']+)["'][^>]*(?:property|name)=["']og:description["']`)
	ogImageRe = regexp.MustCompile(`(?is)<meta[^>]+(?:property|name)=["'](?:og:image|twitter:image)["'][^>]*content=["']([^"']+)["']|<meta[^>]+content=["']([^"']+)["'][^>]*(?:property|name)=["'](?:og:image|twitter:image)["']`)
	titleRe   = regexp.MustCompile(`(?is)<title[^>]*>([^<]+)</title>`)
)

func extractHTTPURL(content string) string {
	m := httpURLRe.FindString(content)
	return strings.TrimRight(m, ".,;:!?")
}

func publicLinkURL(raw string) string {
	raw = strings.TrimSpace(raw)
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return ""
	}
	if hostBlocked(u.Hostname()) {
		return ""
	}
	return raw
}

func ipBlocked(ip net.IP) bool {
	if ip == nil {
		return true
	}
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsMulticast() || ip.IsUnspecified()
}

func hostBlocked(host string) bool {
	h := strings.ToLower(strings.TrimSpace(host))
	if h == "" {
		return true
	}
	if strings.HasPrefix(h, "[") && strings.HasSuffix(h, "]") {
		h = h[1 : len(h)-1]
	} else if ip := net.ParseIP(h); ip == nil {
		if i := strings.LastIndex(h, ":"); i > 0 && !strings.Contains(h, "]") {
			h = h[:i]
		}
	}
	if h == "localhost" || strings.HasSuffix(h, ".localhost") || h == "metadata.google.internal" {
		return true
	}
	if ip := net.ParseIP(h); ip != nil {
		return ipBlocked(ip)
	}
	return false
}

func lookupBlocked(host string) bool {
	if hostBlocked(host) {
		return true
	}
	if net.ParseIP(host) != nil {
		return hostBlocked(host)
	}
	ips, err := net.LookupIP(host)
	if err != nil || len(ips) == 0 {
		return true
	}
	for _, ip := range ips {
		if ipBlocked(ip) {
			return true
		}
	}
	return false
}

func sanitizeImageURL(raw string) string {
	s := strings.TrimSpace(raw)
	if s == "" || utf8.RuneCountInString(s) > 1_200_000 {
		return ""
	}
	if strings.HasPrefix(s, "data:image/") {
		rest := strings.TrimPrefix(s, "data:image/")
		ok := strings.HasPrefix(rest, "png;base64,") ||
			strings.HasPrefix(rest, "jpeg;base64,") ||
			strings.HasPrefix(rest, "jpg;base64,") ||
			strings.HasPrefix(rest, "gif;base64,") ||
			strings.HasPrefix(rest, "webp;base64,")
		if !ok {
			return ""
		}
		return s
	}
	u, err := url.Parse(s)
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" {
		return ""
	}
	if hostBlocked(u.Hostname()) {
		return ""
	}
	if len(s) > 2000 {
		return ""
	}
	return s
}

func firstGroup(m []string) string {
	for i := 1; i < len(m); i++ {
		if m[i] != "" {
			return strings.TrimSpace(htmlUnescape(m[i]))
		}
	}
	return ""
}

func htmlUnescape(s string) string {
	r := strings.NewReplacer("&amp;", "&", "&#39;", "'", "&quot;", `"`, "&lt;", "<", "&gt;", ">")
	return r.Replace(s)
}

func resolveURL(base, ref string) string {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return ""
	}
	u, err := url.Parse(ref)
	if err != nil {
		return ""
	}
	if u.Scheme == "data" {
		return ""
	}
	b, err := url.Parse(base)
	if err != nil {
		return ""
	}
	abs := b.ResolveReference(u)
	if abs.Scheme != "http" && abs.Scheme != "https" {
		return ""
	}
	if hostBlocked(abs.Hostname()) {
		return ""
	}
	return abs.String()
}

func parseLinkMeta(html, pageURL string) (title, desc, image string) {
	if m := ogTitleRe.FindStringSubmatch(html); len(m) > 0 {
		title = firstGroup(m)
	}
	if title == "" {
		if m := titleRe.FindStringSubmatch(html); len(m) > 1 {
			title = htmlUnescape(strings.TrimSpace(m[1]))
		}
	}
	if m := ogDescRe.FindStringSubmatch(html); len(m) > 0 {
		desc = firstGroup(m)
	}
	if m := ogImageRe.FindStringSubmatch(html); len(m) > 0 {
		image = resolveURL(pageURL, firstGroup(m))
	}
	if utf8.RuneCountInString(title) > 120 {
		title = string([]rune(title)[:120])
	}
	if utf8.RuneCountInString(desc) > 240 {
		desc = string([]rune(desc)[:240])
	}
	return title, desc, image
}

func fetchLinkPreview(raw string) (title, desc, image string) {
	raw = strings.TrimSpace(raw)
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return "", "", ""
	}
	if lookupBlocked(u.Hostname()) {
		return "", "", ""
	}
	client := &http.Client{
		Timeout: 4 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 3 {
				return http.ErrUseLastResponse
			}
			if lookupBlocked(req.URL.Hostname()) {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return "", "", ""
	}
	req.Header.Set("User-Agent", "Store4botsLinkPreview/1.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml")
	res, err := client.Do(req)
	if err != nil {
		return "", "", ""
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 400 {
		return "", "", ""
	}
	body, err := io.ReadAll(io.LimitReader(res.Body, 256*1024))
	if err != nil {
		return "", "", ""
	}
	final := raw
	if res.Request != nil && res.Request.URL != nil {
		final = res.Request.URL.String()
	}
	return parseLinkMeta(string(body), final)
}
