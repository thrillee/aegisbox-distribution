package linkscan

import (
	"testing"
)

func normalized(links []Link) []string {
	out := make([]string, len(links))
	for i, l := range links {
		out[i] = l.Normalized
	}
	return out
}

func TestExtractFindsLinks(t *testing.T) {
	cases := []struct {
		name string
		text string
		want []string
	}{
		{"https", "Pay now at https://GTBank-Secure.com/login?id=1#top.", []string{"https://gtbank-secure.com/login?id=1"}},
		{"www", "Visit www.example.com.ng today!", []string{"http://www.example.com.ng/"}},
		{"bare with path", "Claim: bit.ly/3abcXYZ", []string{"http://bit.ly/3abcXYZ"}},
		{"bare common tld", "Go to GTBANK.COM, thanks", []string{"http://gtbank.com/"}},
		{"bare two level suffix", "see opay.com.ng", []string{"http://opay.com.ng/"}},
		{"ip link", "Login http://192.168.1.10:8080/x now", []string{"http://192.168.1.10:8080/x"}},
		{"bare ip with path", "open 45.33.10.2/pay", []string{"http://45.33.10.2/pay"}},
		{"scheme glued to word", "Visit:https://x.ng/a", []string{"https://x.ng/a"}},
		{"wrapped in brackets", "(https://example.com/a)", []string{"https://example.com/a"}},
		{"balanced paren kept", "https://en.wikipedia.org/wiki/Lagos_(state)", []string{"https://en.wikipedia.org/wiki/Lagos_(state)"}},
		{"default port dropped", "https://example.com:443/a", []string{"https://example.com/a"}},
		{"duplicates once", "https://a.com/x and https://A.com/x#y", []string{"https://a.com/x"}},
		{"idn to punycode", "https://bücher.de/x", []string{"https://xn--bcher-kva.de/x"}},
		{"punycode as written", "http://xn--80ak6aa92e.com/", []string{"http://xn--80ak6aa92e.com/"}},
		{"unusual tld with path", "free.love/gift", []string{"http://free.love/gift"}},
		{"whatsapp", "Chat wa.me/2348012345678", []string{"http://wa.me/2348012345678"}},
		{"tel", "Call tel:+2348012345678", []string{"tel:+2348012345678"}},
		{"intent", "intent://scan/#Intent;scheme=zxing;end", []string{"intent://scan/#Intent;scheme=zxing;end"}},
		{"ucs2 text", "Ẹ ṣé! Ṣàbẹ̀wò https://example.ng/ẹ", []string{"https://example.ng/%E1%BA%B9"}},
		{"split across lines", "Link:\nhttps://example.com/a\nbye", []string{"https://example.com/a"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := normalized(Extract(c.text))
			if len(got) != len(c.want) {
				t.Fatalf("Extract(%q) = %v, want %v", c.text, got, c.want)
			}
			for i := range got {
				if got[i] != c.want[i] {
					t.Fatalf("Extract(%q) = %v, want %v", c.text, got, c.want)
				}
			}
		})
	}
}

func TestExtractIgnoresNonLinks(t *testing.T) {
	for _, text := range []string{
		"e.g. this one",
		"You were charged N5.000 today",
		"Dear Mr.Smith, hello",
		"Thanks.Love you",
		"Write to support@example.com",
		"Version 1.2.3 is out",
		"Score 3.14",
		"The end...",
		"com.ng alone",
		"http://",
		"tel:abc",
		"file.txt attached",
	} {
		if got := Extract(text); len(got) != 0 {
			t.Errorf("Extract(%q) = %v, want none", text, normalized(got))
		}
	}
}

func TestExtractCapsLinks(t *testing.T) {
	text := ""
	for i := 0; i < 15; i++ {
		text += " https://example.com/" + string(rune('a'+i))
	}
	if got := Extract(text); len(got) != MaxLinksPerMessage {
		t.Fatalf("got %d links, want %d", len(got), MaxLinksPerMessage)
	}
}

func TestClassify(t *testing.T) {
	cases := []struct {
		text, kind, mode, reason string
	}{
		{"https://example.com/promo", KindWeb, ModeVisit, ""},
		{"https://example.com/unsubscribe?u=1", KindWeb, ModeReputation, ReasonActionLink},
		{"https://example.com/a?token=abc", KindWeb, ModeReputation, ReasonActionLink},
		{"https://example.com/Confirm-Order", KindWeb, ModeReputation, ReasonActionLink},
		{"https://example.com/x?action=opt-out", KindWeb, ModeReputation, ReasonActionLink},
		{"https://wa.me/234801", KindWhatsApp, ModeNone, ReasonNonWeb},
		{"https://chat.whatsapp.com/abc", KindWhatsApp, ModeNone, ReasonNonWeb},
		{"https://play.google.com/store/apps/details?id=x", KindAppStore, ModeNone, ReasonNonWeb},
		{"https://play.google.com/about", KindWeb, ModeVisit, ""},
		{"https://apps.apple.com/app/id1", KindAppStore, ModeNone, ReasonNonWeb},
		{"tel:08012345678", KindTel, ModeNone, ReasonNonWeb},
	}
	for _, c := range cases {
		links := Extract(c.text)
		if len(links) != 1 {
			t.Fatalf("Extract(%q) = %v", c.text, normalized(links))
		}
		l := links[0]
		if l.Kind != c.kind || l.Mode != c.mode || l.Reason != c.reason {
			t.Errorf("%q: got kind=%s mode=%s reason=%s, want %s %s %s", c.text, l.Kind, l.Mode, l.Reason, c.kind, c.mode, c.reason)
		}
	}
}

func TestFlags(t *testing.T) {
	l := Extract("http://10.0.0.1/a")[0]
	if !l.IsIP || l.Host != "10.0.0.1" {
		t.Errorf("ip link: %+v", l)
	}
	l = Extract("https://[::1]:8080/a")[0]
	if !l.IsIP || l.Host != "::1" || l.Normalized != "https://[::1]:8080/a" {
		t.Errorf("ipv6 link: %+v", l)
	}
	l = Extract("https://аррӏе.com/")[0] // Cyrillic look alike
	if !l.IsPunycode {
		t.Errorf("punycode not flagged: %+v", l)
	}
}

func TestGroupKey(t *testing.T) {
	cases := map[string]string{
		"https://promo.example.com/r/a8F3kQ29xZ?u=1":                     "promo.example.com/r/*",
		"https://promo.example.com/r/b7G2mP11yY":                         "promo.example.com/r/*",
		"https://example.com/c/123456789":                                "example.com/c/*",
		"https://example.com/c/550e8400-e29b-41d4-a716-446655440000/pay": "example.com/c/*/pay",
		"https://example.com/sale/shoes":                                 "example.com/sale/shoes",
		"https://bit.ly/3abcXYZ":                                         "bit.ly/3abcXYZ",
	}
	for text, want := range cases {
		if got := Extract(text)[0].GroupKey; got != want {
			t.Errorf("GroupKey(%q) = %q, want %q", text, got, want)
		}
	}
}

func TestLooksLikeOTP(t *testing.T) {
	yes := []string{
		"Your OTP is 482913. Do not share it.",
		"Your verification code is 1234",
		"123456 is your one-time password",
		"Your login code is 99887766",
	}
	no := []string{
		"Big sale! 50% off at https://shop.ng",
		"Your OTP has expired, request a new one",
		"Your order 123456 has shipped",
	}
	for _, s := range yes {
		if !LooksLikeOTP(s) {
			t.Errorf("LooksLikeOTP(%q) = false", s)
		}
	}
	for _, s := range no {
		if LooksLikeOTP(s) {
			t.Errorf("LooksLikeOTP(%q) = true", s)
		}
	}
}
