package smppserver

import "testing"

func TestReceiptErrorCode(t *testing.T) {
	s := func(v string) *string { return &v }
	cases := []struct {
		status string
		code   *string
		want   string
	}{
		{"DELIVRD", s("000"), "000"},
		{"DELIVRD", s("008"), "000"}, // delivered is never an error
		{"DELIVRD", nil, "000"},
		{"UNDELIV", s("001"), "001"}, // upstream SMPP code passes through
		{"UNDELIV", s("0bf"), "0BF"},
		{"UNDELIV", nil, "000"},
		{"UNDELIV", s(""), "000"},
	}
	for _, c := range cases {
		if got := receiptErrorCode(c.status, c.code); got != c.want {
			t.Errorf("receiptErrorCode(%s, %v) = %s, want %s", c.status, c.code, got, c.want)
		}
	}
	// Aegisbox's own longer codes still go through the mapper (default 008 when unknown).
	if got := receiptErrorCode("UNDELIV", s("SOME_INTERNAL_CODE")); len(got) != 3 {
		t.Errorf("internal code mapped to %q, want a 3 character SMPP code", got)
	}
}
