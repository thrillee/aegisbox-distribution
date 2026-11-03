// Package linkscan finds links in message text and queues them for the link scanner.
//
// Only messages sent through a simulator MNO are scanned (see LINK_SCANNING_PLAN.md in
// aegisbox-distribution). This file holds the extraction rules; enqueue.go writes the queue.
package linkscan

import (
	"net"
	"net/url"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/net/idna"
	"golang.org/x/net/publicsuffix"
)

// Link kinds.
const (
	KindWeb      = "web"
	KindWhatsApp = "whatsapp"
	KindTel      = "tel"
	KindIntent   = "intent"
	KindAppStore = "app_store"
)

// Scan modes.
const (
	ModeVisit      = "visit"      // open in the scanner's browser
	ModeReputation = "reputation" // reputation and domain checks only, never opened
	ModeNone       = "none"       // recorded only
)

// Reasons a link is not visited.
const (
	ReasonNonWeb     = "non_web"
	ReasonActionLink = "action_link"
	ReasonSampled    = "sampled"
)

// MaxLinksPerMessage caps how many links one message can queue.
const MaxLinksPerMessage = 10

// Link is one link found in a message.
type Link struct {
	Raw        string // as written in the message
	Normalized string // scheme added, host lowercased and in punycode, no fragment
	Host       string // lowercased punycode host (or the number for tel:)
	Kind       string
	Mode       string
	Reason     string // set when Mode is not ModeVisit
	GroupKey   string // host and path with per recipient tokens replaced by *
	IsIP       bool
	IsPunycode bool

	// Filled by the scan worker before scoring, not by extraction.
	AgeDomain     string // registrable domain whose age was looked up
	DomainAgeDays *int   // days since registration (RDAP); nil when unknown
	ListedBy      string // reputation list that has the URL (e.g. "URLhaus: malware_download")
}

var (
	schemeRe = regexp.MustCompile(`(?i)(https?://|intent://|tel:)`)
	// Labels: letters (any script, for IDN), digits and inner hyphens.
	labelRe = regexp.MustCompile(`^[\p{L}\p{N}](?:[\p{L}\p{N}-]*[\p{L}\p{N}])?$`)
	telRe   = regexp.MustCompile(`^\+?[0-9][0-9\-\s().]{3,}$`)
)

// commonBareTLDs are suffixes accepted for a bare domain written without a path ("gtbank.com").
// Other listed suffixes need a path ("x.love/a"), so ordinary words joined by a dot
// ("Thanks.Love") are not taken for links.
var commonBareTLDs = map[string]bool{
	"com": true, "ng": true, "net": true, "org": true, "io": true, "co": true, "info": true,
	"biz": true, "xyz": true, "link": true, "site": true, "online": true, "app": true,
	"me": true, "ly": true, "gl": true, "to": true, "cc": true, "top": true, "club": true,
	"shop": true, "store": true, "live": true, "click": true, "vip": true, "win": true,
	"ru": true, "cn": true, "tk": true, "ml": true, "ga": true, "cf": true, "gq": true,
	"uk": true, "us": true, "africa": true, "page": true, "dev": true, "ai": true,
}

// shortenerHosts are link shorteners. Their links are not grouped as per recipient links: each
// short code is its own link (the scanner expands them through preview endpoints, step 3).
var shortenerHosts = map[string]bool{
	"bit.ly": true, "bitly.com": true, "tinyurl.com": true, "t.co": true, "goo.gl": true,
	"ow.ly": true, "is.gd": true, "buff.ly": true, "cutt.ly": true, "rb.gy": true,
	"shorturl.at": true, "t.ly": true, "rebrand.ly": true, "tiny.cc": true, "s.id": true,
	"bit.do": true, "lnkd.in": true, "short.io": true, "v.gd": true, "shorte.st": true,
}

var whatsAppHosts = map[string]bool{
	"wa.me": true, "whatsapp.com": true, "api.whatsapp.com": true, "chat.whatsapp.com": true,
	"web.whatsapp.com": true, "wa.link": true,
}

// actionWords in a path or query mark a link whose plain GET may do something (unsubscribe,
// confirm an order): it is checked by reputation only, never opened.
var actionWords = []string{"unsubscribe", "optout", "opt-out", "opt_out", "confirm", "verify", "approve"}

// tokenParams are query parameters that usually carry a one time token.
var tokenParams = map[string]bool{
	"token": true, "otp": true, "auth": true, "auth_token": true, "access_token": true,
	"sig": true, "signature": true, "magic": true, "login_token": true,
}

// Extract returns the distinct links in text, in order of appearance, at most MaxLinksPerMessage.
// Pass the full message text, never single segments, so a link split across parts is found.
// Email addresses are not links.
func Extract(text string) []Link {
	var out []Link
	seen := map[string]bool{}
	for _, tok := range splitTokens(text) {
		l, ok := parseToken(tok)
		if !ok || seen[l.Normalized] {
			continue
		}
		seen[l.Normalized] = true
		out = append(out, l)
		if len(out) == MaxLinksPerMessage {
			break
		}
	}
	return out
}

// splitTokens splits on whitespace and on characters that never appear inside a link as written
// in an SMS (quotes, angle brackets, some unicode punctuation).
func splitTokens(text string) []string {
	return strings.FieldsFunc(text, func(r rune) bool {
		if unicode.IsSpace(r) {
			return true
		}
		switch r {
		case '"', '<', '>', '`', '|', '“', '”', '‘', '’', '«', '»', '。', '，':
			return true
		}
		return false
	})
}

// trimToken strips punctuation that ends a sentence or wraps a link. A closing bracket is kept
// when the token also contains its opening bracket ("en.wikipedia.org/wiki/Lagos_(state)").
func trimToken(tok string) string {
	tok = strings.TrimLeft(tok, "([{'*_")
	for len(tok) > 0 {
		last := tok[len(tok)-1]
		switch {
		case strings.IndexByte(".,;:!?'*_", last) >= 0:
			tok = tok[:len(tok)-1]
		case last == ')' && !strings.Contains(tok, "("),
			last == ']' && !strings.Contains(tok, "["),
			last == '}' && !strings.Contains(tok, "{"):
			tok = tok[:len(tok)-1]
		default:
			return tok
		}
	}
	return tok
}

func parseToken(tok string) (Link, bool) {
	tok = trimToken(tok)
	if tok == "" {
		return Link{}, false
	}
	// A scheme anywhere in the token ("Visit:https://x.ng") starts the link there.
	if loc := schemeRe.FindStringIndex(tok); loc != nil {
		raw := trimToken(tok[loc[0]:])
		scheme := strings.ToLower(tok[loc[0]:loc[1]])
		switch scheme {
		case "tel:":
			num := strings.TrimPrefix(strings.ToLower(raw), "tel:")
			num = strings.TrimPrefix(num, "//")
			if !telRe.MatchString(num) {
				return Link{}, false
			}
			return Link{Raw: raw, Normalized: "tel:" + num, Host: num, Kind: KindTel,
				Mode: ModeNone, Reason: ReasonNonWeb, GroupKey: "tel:" + num}, true
		case "intent://":
			norm := "intent://" + raw[len(scheme):]
			host := raw[len(scheme):]
			if i := strings.IndexAny(host, "/#?"); i >= 0 {
				host = host[:i]
			}
			return Link{Raw: raw, Normalized: norm, Host: strings.ToLower(host), Kind: KindIntent,
				Mode: ModeNone, Reason: ReasonNonWeb, GroupKey: "intent://" + strings.ToLower(host)}, true
		default:
			return parseWeb(raw, true)
		}
	}
	if strings.Contains(tok, "@") { // email address (or user@host), not a link
		return Link{}, false
	}
	if !strings.Contains(tok, ".") {
		return Link{}, false
	}
	return parseWeb(tok, false)
}

// parseWeb checks and normalizes an http(s) link or a bare domain such as "www.x.ng" or
// "bit.ly/abc".
func parseWeb(raw string, hasScheme bool) (Link, bool) {
	s := raw
	if !hasScheme {
		s = "http://" + raw
	}
	u, err := url.Parse(s)
	if err != nil || u.Host == "" || u.User != nil {
		return Link{}, false
	}
	host := strings.TrimSuffix(strings.ToLower(u.Hostname()), ".")
	if host == "" {
		return Link{}, false
	}

	l := Link{Raw: raw, Kind: KindWeb}
	if ip := net.ParseIP(host); ip != nil {
		l.IsIP = true
		if ip.To4() == nil {
			host = "[" + ip.String() + "]"
		} else {
			host = ip.String()
		}
	} else {
		ascii, ok := checkDomain(host, u, hasScheme)
		if !ok {
			return Link{}, false
		}
		host = ascii
		l.IsPunycode = strings.HasPrefix(host, "xn--") || strings.Contains(host, ".xn--")
	}

	u.Scheme = strings.ToLower(u.Scheme)
	u.Fragment = ""
	u.RawFragment = ""
	port := u.Port()
	if (u.Scheme == "http" && port == "80") || (u.Scheme == "https" && port == "443") {
		port = ""
	}
	u.Host = host
	if port != "" {
		u.Host = host + ":" + port
	}
	if u.Path == "" && u.RawQuery == "" {
		u.Path = "/"
	}
	l.Normalized = u.String()
	l.Host = strings.Trim(host, "[]")
	l.GroupKey = groupKey(l.Host, u)
	l.Kind, l.Mode, l.Reason = classify(l.Host, u)
	return l, true
}

// checkDomain validates a host name and returns it in punycode. A host is accepted only when its
// suffix is in the public suffix list, so "e.g.", "N5.000" and "Mr.Smith" are not links.
func checkDomain(host string, u *url.URL, hasScheme bool) (string, bool) {
	ascii, err := idna.Lookup.ToASCII(host)
	if err != nil || !strings.Contains(ascii, ".") {
		return "", false
	}
	labels := strings.Split(ascii, ".")
	for _, lab := range labels {
		if lab == "" || len(lab) > 63 || !labelRe.MatchString(lab) {
			return "", false
		}
	}
	suffix, icann := publicsuffix.PublicSuffix(ascii)
	// An unlisted suffix comes back as the last label with icann false; listed private suffixes
	// (github.io) contain a dot.
	if !icann && !strings.Contains(suffix, ".") {
		return "", false
	}
	if suffix == ascii { // the host is only a suffix ("com.ng")
		return "", false
	}
	if hasScheme || strings.HasPrefix(ascii, "www.") {
		return ascii, true
	}
	// A bare domain without a path must use a common suffix, or a two level one like "com.ng".
	hasPath := (u.Path != "" && u.Path != "/") || u.RawQuery != ""
	if hasPath || commonBareTLDs[suffix] || strings.Contains(suffix, ".") {
		return ascii, true
	}
	return "", false
}

func classify(host string, u *url.URL) (kind, mode, reason string) {
	bare := strings.TrimPrefix(host, "www.")
	switch {
	case whatsAppHosts[bare]:
		return KindWhatsApp, ModeNone, ReasonNonWeb
	case bare == "play.google.com" && strings.HasPrefix(u.Path, "/store"),
		bare == "apps.apple.com", bare == "itunes.apple.com":
		return KindAppStore, ModeNone, ReasonNonWeb
	}
	if isActionLink(u) {
		return KindWeb, ModeReputation, ReasonActionLink
	}
	return KindWeb, ModeVisit, ""
}

func isActionLink(u *url.URL) bool {
	lowerPath := strings.ToLower(u.EscapedPath())
	lowerQuery := strings.ToLower(u.RawQuery)
	for _, w := range actionWords {
		if strings.Contains(lowerPath, w) || strings.Contains(lowerQuery, w) {
			return true
		}
	}
	for k := range u.Query() {
		if tokenParams[strings.ToLower(k)] {
			return true
		}
	}
	return false
}

// groupKey groups links that differ only by a per recipient token, so only a sample of them is
// visited: host plus path, with token like path segments replaced by * and the query dropped.
// Shortener links are not grouped.
func groupKey(host string, u *url.URL) string {
	if shortenerHosts[strings.TrimPrefix(host, "www.")] {
		return host + u.EscapedPath()
	}
	segs := strings.Split(u.EscapedPath(), "/")
	for i, seg := range segs {
		if looksLikeToken(seg) {
			segs[i] = "*"
		}
	}
	return host + strings.Join(segs, "/")
}

var (
	hexRe  = regexp.MustCompile(`^[0-9a-fA-F]{8,}$`)
	uuidRe = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	digRe  = regexp.MustCompile(`^[0-9]{6,}$`)
)

func looksLikeToken(seg string) bool {
	if len(seg) < 6 {
		return false
	}
	if hexRe.MatchString(seg) || uuidRe.MatchString(seg) || digRe.MatchString(seg) {
		return true
	}
	var letters, digits bool
	for _, r := range seg {
		switch {
		case r >= '0' && r <= '9':
			digits = true
		case unicode.IsLetter(r):
			letters = true
		case r == '-' || r == '_':
		default:
			return false
		}
	}
	return letters && digits
}

var (
	otpWordRe = regexp.MustCompile(`(?i)\b(otp|one[\s-]?time|verification code|verify code|auth(?:entication)? code|security code|passcode|pass code|login code|your code|code is|pin is|token is)\b`)
	otpNumRe  = regexp.MustCompile(`\b\d{4,8}\b`)
)

// LooksLikeOTP reports whether text reads like a one time password message: an OTP phrase and a 4
// to 8 digit number.
func LooksLikeOTP(text string) bool {
	return otpWordRe.MatchString(text) && otpNumRe.MatchString(text)
}
