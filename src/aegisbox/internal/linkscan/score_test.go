package linkscan

import (
	"strings"
	"testing"
)

func sp(s string) *string { return &s }
func ip(i int) *int       { return &i }

func link(t *testing.T, raw string) Link {
	t.Helper()
	ls := Extract(raw)
	if len(ls) != 1 {
		t.Fatalf("Extract(%q) = %v", raw, ls)
	}
	return ls[0]
}

func codes(rs []Reason) string {
	var c []string
	for _, r := range rs {
		c = append(c, r.Code)
	}
	return strings.Join(c, ",")
}

func TestScoreSafePage(t *testing.T) {
	v, s, r := Score(link(t, "https://shop.example.com/sale"), &ScanResult{
		FinalURL: sp("https://shop.example.com/sale"), Status: ip(200), Title: sp("Shoe sale"),
		Redirects: []Redirect{{URL: "https://shop.example.com/sale", Type: "start"}},
	})
	if v != VerdictSafe || s != 0 {
		t.Fatalf("got %s %d %s", v, s, codes(r))
	}
}

func TestScoreFakeBankLogin(t *testing.T) {
	v, s, r := Score(link(t, "http://gtbank-secure-login.com/verify"), &ScanResult{
		FinalURL: sp("http://gtbank-secure-login.com/verify"), Status: ip(200),
		Title: sp("GTBank | Secure Login"), Text: sp("Your account is restricted"),
		Forms:         []Form{{Fields: []Field{{Type: "password", Name: "password"}, {Name: "card_number"}, {Name: "pin"}}}},
		BlockedWrites: 1,
	})
	if v != VerdictMalicious {
		t.Fatalf("got %s %d %s", v, s, codes(r))
	}
	for _, c := range []string{"brand_in_domain", "asks_for_secrets", "no_https", "brand_on_page", "auto_submit"} {
		if !strings.Contains(codes(r), c) {
			t.Errorf("missing %s in %s", c, codes(r))
		}
	}
}

func TestScoreRealBankLoginIsNotFlagged(t *testing.T) {
	v, s, r := Score(link(t, "https://ibank.gtbank.com/login"), &ScanResult{
		FinalURL: sp("https://ibank.gtbank.com/login"), Status: ip(200), Title: sp("GTBank Internet Banking"),
		Forms: []Form{{Fields: []Field{{Type: "password", Name: "password"}, {Name: "pin"}}}},
	})
	if v != VerdictSafe {
		t.Fatalf("got %s %d %s", v, s, codes(r))
	}
}

func TestScoreAPK(t *testing.T) {
	v, _, r := Score(link(t, "https://get-app.xyz/download"), &ScanResult{
		FinalURL: sp("https://get-app.xyz/download"), Status: ip(200),
		Downloads: []Download{{Filename: "update.apk", Mime: "application/vnd.android.package-archive"}},
	})
	if v != VerdictSuspicious || !strings.Contains(codes(r), "apk_download") {
		t.Fatalf("got %s %s", v, codes(r))
	}
}

func TestScoreGeoBlockedIsNeverSafe(t *testing.T) {
	v, _, r := Score(link(t, "https://promo.example.com/"), &ScanResult{
		FinalURL: sp("https://promo.example.com/"), Status: ip(403), Text: sp("Forbidden"),
	})
	if v != VerdictSuspicious || !strings.Contains(codes(r), "may_hide_outside_nigeria") {
		t.Fatalf("got %s %s", v, codes(r))
	}
}

func TestScoreInternalAddress(t *testing.T) {
	v, s, r := Score(link(t, "http://10.0.0.1/admin"), &ScanResult{
		ErrorKind: sp("blocked"), Error: sp("blocked"), Blocked: []Blocked{{Host: "10.0.0.1", IP: sp("10.0.0.1"), Reason: "private"}},
	})
	if v != VerdictSuspicious || s != 65 { // ip_link 25 + internal_address 40, counted once
		t.Fatalf("got %s %d %s", v, s, codes(r))
	}
}

func TestScoreRedirectToOtherDomainAndShortener(t *testing.T) {
	_, _, r := Score(link(t, "https://news.example.com/a"), &ScanResult{
		FinalURL: sp("https://other.example.net/b"), Status: ip(200),
		Redirects: []Redirect{{Type: "start"}, {Type: "http"}, {Type: "http"}, {Type: "script_or_meta"}, {Type: "http"}},
	})
	if !strings.Contains(codes(r), "domain_changed") || !strings.Contains(codes(r), "many_redirects") {
		t.Fatalf("got %s", codes(r))
	}
	_, _, r = Score(link(t, "https://bit.ly/abc123"), &ScanResult{
		FinalURL: sp("https://shop.example.com/"), Status: ip(200), ShortenerExpanded: true,
		Redirects: []Redirect{{Type: "shortener"}, {Type: "start"}},
	})
	if strings.Contains(codes(r), "domain_changed") {
		t.Fatalf("a shortener leading elsewhere is expected: %s", codes(r))
	}
}

func TestScoreStatic(t *testing.T) {
	v, _, r := Score(link(t, "https://gtbänk.com/verify"), nil) // stored as punycode
	if v == VerdictSafe || !strings.Contains(codes(r), "punycode") || !strings.Contains(codes(r), "brand_in_domain") {
		t.Fatalf("got %s %s", v, codes(r))
	}
	v, s, _ := Score(link(t, "https://example.com/unsubscribe"), nil)
	if v != VerdictSafe || s != 0 {
		t.Fatalf("got %s %d", v, s)
	}
}

func TestImpersonatedBrand(t *testing.T) {
	cases := map[string]string{
		"gtbank-secure-login.com": "GTBank",
		"www.gtbank.com":          "",
		"opay-bonus.top":          "OPay",
		"mtn.ng":                  "",
		"mtn-free-data.xyz":       "MTN",
		"dubai-trip.com":          "",
		"cuba.example.com":        "",
		"uba-verify.com":          "UBA",
		"zenithbankonline.net":    "Zenith Bank",
		"shop.example.com":        "",
	}
	for host, want := range cases {
		if got := impersonatedBrand(host); got != want {
			t.Errorf("impersonatedBrand(%q) = %q, want %q", host, got, want)
		}
	}
}
