package linkscan

import (
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strings"

	"golang.org/x/net/idna"
	"golang.org/x/net/publicsuffix"
	"golang.org/x/text/unicode/norm"
)

// Verdicts.
const (
	VerdictSafe       = "safe"
	VerdictSuspicious = "suspicious"
	VerdictMalicious  = "malicious"
)

// Score bands: malicious 70 and above, suspicious 30 to 69, safe below 30.
const (
	maliciousScore  = 70
	suspiciousScore = 30
)

// Reason is one signal that added to a link's score.
type Reason struct {
	Code   string `json:"code"`
	Points int    `json:"points"`
	Detail string `json:"detail"`
}

// ScanResult is what the linkscanner service reports for one URL.
type ScanResult struct {
	URL               string     `json:"url"`
	FinalURL          *string    `json:"finalUrl"`
	Status            *int       `json:"status"`
	Title             *string    `json:"title"`
	Text              *string    `json:"text"`
	Redirects         []Redirect `json:"redirects"`
	Forms             []Form     `json:"forms"`
	Downloads         []Download `json:"downloads"`
	Blocked           []Blocked  `json:"blocked"`
	TLSIssuer         *string    `json:"tlsIssuer"`
	Screenshot        *string    `json:"screenshot"` // base64 JPEG
	Error             *string    `json:"error"`
	ErrorKind         *string    `json:"errorKind"`
	ShortenerExpanded bool       `json:"shortenerExpanded"`
	BlockedWrites     int        `json:"blockedWrites"`
	DurationMs        int        `json:"durationMs"`
}

type Redirect struct {
	URL    string `json:"url"`
	Status *int   `json:"status"`
	Type   string `json:"type"`
}

type Form struct {
	Action string  `json:"action"`
	Method string  `json:"method"`
	Fields []Field `json:"fields"`
}

type Field struct {
	Tag          string `json:"tag"`
	Type         string `json:"type"`
	Name         string `json:"name"`
	ID           string `json:"id"`
	Placeholder  string `json:"placeholder"`
	Autocomplete string `json:"autocomplete"`
	Label        string `json:"label"`
}

type Download struct {
	URL      string `json:"url"`
	Filename string `json:"filename"`
	Mime     string `json:"mime"`
}

type Blocked struct {
	Host   string  `json:"host"`
	IP     *string `json:"ip"`
	Reason string  `json:"reason"`
}

// brand is a Nigerian bank, fintech or telco whose name on someone else's domain suggests
// impersonation.
type brand struct {
	name    string
	tokens  []string // host words that name the brand
	domains []string // registrable domains the brand owns
}

var brands = []brand{
	{"GTBank", []string{"gtbank", "gtco", "guarantytrust"}, []string{"gtbank.com", "gtco.com", "gtbank.ng"}},
	{"Access Bank", []string{"accessbank", "accessbankplc"}, []string{"accessbankplc.com", "accessbank.ng"}},
	{"Zenith Bank", []string{"zenithbank", "zenith"}, []string{"zenithbank.com"}},
	{"First Bank", []string{"firstbank", "firstbanknigeria", "firstbanknig"}, []string{"firstbanknigeria.com", "firstbank.ng"}},
	{"UBA", []string{"uba", "ubagroup", "ubaonline"}, []string{"ubagroup.com", "uba.ng"}},
	{"OPay", []string{"opay", "opayweb"}, []string{"opayweb.com", "opay.ng", "opaycheckout.com"}},
	{"PalmPay", []string{"palmpay"}, []string{"palmpay.com", "palmpay.ng"}},
	{"Moniepoint", []string{"moniepoint"}, []string{"moniepoint.com"}},
	{"Kuda", []string{"kuda", "kudabank"}, []string{"kuda.com", "kudabank.com"}},
	{"MTN", []string{"mtn", "mtnng", "mtnnigeria"}, []string{"mtn.ng", "mtn.com", "mtnonline.com"}},
	{"Airtel", []string{"airtel", "airtelng"}, []string{"airtel.com.ng", "airtel.com", "airtel.africa"}},
}

// unusualTLDs are suffixes common in throwaway phishing domains.
var unusualTLDs = map[string]bool{
	"zip": true, "mov": true, "xyz": true, "top": true, "click": true, "link": true, "live": true,
	"icu": true, "cyou": true, "sbs": true, "rest": true, "tk": true, "ml": true, "ga": true,
	"cf": true, "gq": true, "buzz": true, "monster": true, "quest": true, "cfd": true, "bond": true,
	"loan": true, "win": true, "work": true, "support": true, "vip": true, "shop": true,
}

var (
	cardFieldRe   = regexp.MustCompile(`(?i)card.?(num|no)|cc.?num|\bpan\b|cvv|cvc|security code|expir|exp.?date`)
	pinFieldRe    = regexp.MustCompile(`(?i)\bpin\b|atm.?pin|pin.?code|passcode`)
	otpFieldRe    = regexp.MustCompile(`(?i)\botp\b|one.?time|token|verification code`)
	idFieldRe     = regexp.MustCompile(`(?i)\bbvn\b|\bnin\b|national id`)
	geoTextRe     = regexp.MustCompile(`(?i)not available in your (country|region|location)|access denied.*(country|region)|unavailable in your (country|region)`)
	brandTextRe   = regexp.MustCompile(`(?i)\b(gtbank|gtco|access bank|zenith bank|first ?bank|\buba\b|opay|palmpay|moniepoint|kuda|mtn|airtel)\b`)
	riskyDownload = regexp.MustCompile(`(?i)\.(apk|xapk|exe|msi|scr|bat|cmd|js|vbs|jar|dmg|pkg|ipa)$`)
)

// registrable returns the registrable domain (example.com.ng) of a host, or the host itself.
func registrable(host string) string {
	if net.ParseIP(host) != nil {
		return host
	}
	if d, err := publicsuffix.EffectiveTLDPlusOne(host); err == nil {
		return d
	}
	return host
}

func hostWords(host string) []string {
	return strings.FieldsFunc(strings.ToLower(host), func(r rune) bool {
		return !(r >= 'a' && r <= 'z')
	})
}

// confusables maps common Cyrillic and Greek look alike letters to Latin ones.
var confusables = map[rune]rune{
	'а': 'a', 'е': 'e', 'о': 'o', 'р': 'p', 'с': 'c', 'у': 'y', 'х': 'x', 'і': 'i', 'ӏ': 'l',
	'ԁ': 'd', 'ѕ': 's', 'ј': 'j', 'һ': 'h', 'к': 'k', 'м': 'm', 'т': 't', 'в': 'b', 'п': 'n',
	'α': 'a', 'ο': 'o', 'ρ': 'p', 'τ': 't', 'κ': 'k', 'ν': 'v', 'ι': 'i', 'μ': 'u',
}

// skeleton turns a punycode host into plain Latin letters for brand matching: "xn--gtbnk-qta.com"
// (gtbänk.com) becomes "gtbank.com".
func skeleton(host string) string {
	u, err := idna.ToUnicode(host)
	if err != nil {
		return host
	}
	var b strings.Builder
	for _, r := range norm.NFD.String(u) {
		if r >= 0x300 && r <= 0x36f { // combining marks (accents)
			continue
		}
		if c, ok := confusables[r]; ok {
			r = c
		}
		b.WriteRune(r)
	}
	return b.String()
}

// impersonatedBrand returns the brand named in host when host is not one of its domains.
func impersonatedBrand(host string) string {
	reg := registrable(host)
	words := hostWords(strings.TrimSuffix(host, "."))
	if strings.Contains(host, "xn--") {
		words = append(words, hostWords(skeleton(host))...)
	}
	for _, b := range brands {
		owned := false
		for _, d := range b.domains {
			if reg == d {
				owned = true
				break
			}
		}
		if owned {
			continue
		}
		for _, w := range words {
			for _, t := range b.tokens {
				// Short names (uba, mtn, kuda) must be a whole word; longer ones may be part of one.
				if w == t || (len(t) >= 5 && strings.Contains(w, t)) {
					return b.name
				}
			}
		}
	}
	return ""
}

func ownsBrandDomain(host string) bool {
	reg := registrable(host)
	for _, b := range brands {
		for _, d := range b.domains {
			if reg == d {
				return true
			}
		}
	}
	return false
}

func suffixOf(host string) string {
	s, _ := publicsuffix.PublicSuffix(host)
	return s
}

func strOr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// Score judges a link. res is nil for a static check (the link was not opened).
// geoUnsure is set when the page may show different content outside Nigeria; such a link is
// never judged safe.
func Score(l Link, res *ScanResult) (verdict string, score int, reasons []Reason) {
	add := func(code string, points int, format string, a ...any) {
		reasons = append(reasons, Reason{Code: code, Points: points, Detail: fmt.Sprintf(format, a...)})
		score += points
	}

	// Signals from the link as written.
	if l.IsIP {
		add("ip_link", 25, "link uses an IP address (%s) instead of a name", l.Host)
	}
	if l.IsPunycode {
		add("punycode", 25, "link uses an internationalized (punycode) domain %s, often used for look alike names", l.Host)
	}
	if b := impersonatedBrand(l.Host); b != "" {
		add("brand_in_domain", 40, "domain %s names %s but is not one of its domains", l.Host, b)
	}
	if t := suffixOf(l.Host); unusualTLDs[t] {
		add("unusual_tld", 15, "domain ends in .%s, common for throwaway sites", t)
	}

	geoUnsure := false
	if res != nil {
		finalURL := strOr(res.FinalURL)
		finalHost := ""
		if u, err := url.Parse(finalURL); err == nil {
			finalHost = strings.ToLower(u.Hostname())
			if u.Scheme == "http" {
				add("no_https", 10, "page is served without HTTPS")
			}
		}
		if finalHost != "" && finalHost != l.Host {
			if b := impersonatedBrand(finalHost); b != "" && impersonatedBrand(l.Host) == "" {
				add("brand_in_domain", 40, "link leads to %s, which names %s but is not one of its domains", finalHost, b)
			}
			if net.ParseIP(finalHost) != nil && !l.IsIP {
				add("ip_link", 25, "link leads to an IP address (%s)", finalHost)
			}
			if registrable(finalHost) != registrable(l.Host) && !res.ShortenerExpanded {
				add("domain_changed", 10, "link shows %s but leads to %s", registrable(l.Host), registrable(finalHost))
			}
		}

		hops := 0
		for _, r := range res.Redirects {
			if r.Type != "start" {
				hops++
			}
		}
		if hops > 3 {
			add("many_redirects", 10, "%d redirects before the final page", hops)
		}

		for _, b := range res.Blocked {
			add("internal_address", 40, "page tried to reach a private or internal address (%s: %s)", b.Host, b.Reason)
			break
		}

		// Forms asking for secrets.
		var hasPassword, hasCard, hasPin, hasOTP, hasID bool
		for _, f := range res.Forms {
			for _, fd := range f.Fields {
				text := strings.Join([]string{fd.Name, fd.ID, fd.Placeholder, fd.Autocomplete, fd.Label}, " ")
				switch {
				case fd.Type == "password":
					hasPassword = true
				}
				if cardFieldRe.MatchString(text) || strings.HasPrefix(fd.Autocomplete, "cc-") {
					hasCard = true
				}
				if pinFieldRe.MatchString(text) {
					hasPin = true
				}
				if otpFieldRe.MatchString(text) || fd.Autocomplete == "one-time-code" {
					hasOTP = true
				}
				if idFieldRe.MatchString(text) {
					hasID = true
				}
			}
		}
		asks := []string{}
		points := 0
		if hasCard {
			asks, points = append(asks, "card details"), points+35
		}
		if hasPin {
			asks, points = append(asks, "PIN"), points+35
		}
		if hasID {
			asks, points = append(asks, "BVN or NIN"), points+30
		}
		if hasOTP {
			asks, points = append(asks, "OTP"), points+25
		}
		if hasPassword {
			asks, points = append(asks, "password"), points+20
		}
		if points > 0 {
			if points > 60 {
				points = 60
			}
			if ownsBrandDomain(finalHost) {
				// The real bank's own login page is expected to ask for these.
				points = 0
			}
			if points > 0 {
				add("asks_for_secrets", points, "page asks for %s", strings.Join(asks, ", "))
			}
		}
		pageText := strOr(res.Title) + " " + strOr(res.Text)
		if points > 0 && finalHost != "" && !ownsBrandDomain(finalHost) {
			if m := brandTextRe.FindString(pageText); m != "" {
				add("brand_on_page", 30, "page uses the name %q and asks for secrets on %s", m, registrable(finalHost))
			}
		}
		if res.BlockedWrites > 0 && points > 0 {
			add("auto_submit", 10, "page tried to send data without anyone clicking (%d attempts blocked)", res.BlockedWrites)
		}

		for _, d := range res.Downloads {
			name := strings.ToLower(d.Filename + " " + d.URL)
			if strings.Contains(name, ".apk") || strings.Contains(d.Mime, "android.package-archive") {
				add("apk_download", 40, "page starts an Android app download (%s), outside the app stores", d.Filename)
			} else if riskyDownload.MatchString(strings.ToLower(d.Filename)) {
				add("risky_download", 30, "page starts a download of %s", d.Filename)
			} else {
				add("download", 10, "page starts a download (%s)", d.Filename)
			}
			break
		}

		switch strOr(res.ErrorKind) {
		case "dns":
			add("unreachable", 10, "domain does not resolve (%s)", strOr(res.Error))
		case "timeout":
			add("unreachable", 10, "page did not load in time")
		case "tls":
			add("bad_certificate", 20, "HTTPS certificate is not valid (%s)", strOr(res.Error))
		case "redirect_loop":
			add("redirect_loop", 15, "redirect loop")
		case "blocked":
			if len(res.Blocked) == 0 {
				add("internal_address", 40, "link points to a private or internal address")
			}
		case "size_limit", "other":
			add("unreachable", 5, "page could not be checked (%s)", strOr(res.Error))
		}

		status := 0
		if res.Status != nil {
			status = *res.Status
		}
		if status >= 400 && status != 403 && status != 451 && strOr(res.ErrorKind) == "" {
			add("page_error", 5, "page answered with HTTP %d", status)
		}
		if status == 403 || status == 451 || geoTextRe.MatchString(pageText) {
			geoUnsure = true
			add("may_hide_outside_nigeria", 0, "page refused the check (HTTP %d) or says it is not available here; it may show different content in Nigeria", status)
		}
	}

	// Domain age (RDAP) and reputation lists, looked up by the worker.
	if l.DomainAgeDays != nil {
		switch d := *l.DomainAgeDays; {
		case d < 30:
			add("new_domain", 20, "domain %s was registered %d days ago", l.AgeDomain, d)
		case d < 90:
			add("young_domain", 10, "domain %s was registered %d days ago", l.AgeDomain, d)
		}
	}
	if l.ListedBy != "" {
		add("listed_malicious", 70, "URL is on a malicious URL list (%s)", l.ListedBy)
	}

	switch {
	case score >= maliciousScore:
		verdict = VerdictMalicious
	case score >= suspiciousScore:
		verdict = VerdictSuspicious
	default:
		verdict = VerdictSafe
	}
	if geoUnsure && verdict == VerdictSafe {
		verdict = VerdictSuspicious
	}
	if reasons == nil {
		reasons = []Reason{}
	}
	return verdict, score, reasons
}
