// Package mnosim is a simulated MNO: an SMPP 3.4 server that accepts binds and submits like an
// upstream operator, answers each submit with a message ID and later sends a delivery receipt.
// It is for test traffic only; the gateway refuses to route customers not marked as test to an
// MNO flagged as a simulator.
package mnosim

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config is read from MNOSIM_* environment variables.
type Config struct {
	Addr     string `envconfig:"MNOSIM_ADDR" default:":2776"`
	SystemID string `envconfig:"MNOSIM_SYSTEM_ID" default:"simulator"`
	Password string `envconfig:"MNOSIM_PASSWORD" default:"simpass"`

	// DLR delay range after a submit is accepted.
	DLRDelayMin time.Duration `envconfig:"MNOSIM_DLR_DELAY_MIN" default:"2s"`
	DLRDelayMax time.Duration `envconfig:"MNOSIM_DLR_DELAY_MAX" default:"10s"`
	// Outcome mix as STATUS:WEIGHT pairs; NONE means no DLR is ever sent.
	DLRMix string `envconfig:"MNOSIM_DLR_MIX" default:"DELIVRD:90,UNDELIV:5,EXPIRED:2,NONE:3"`

	// Share (0 to 100) of submits rejected with RejectStatus, and of submits never answered.
	RejectPercent     float64 `envconfig:"MNOSIM_REJECT_PERCENT" default:"0"`
	RejectStatus      uint32  `envconfig:"MNOSIM_REJECT_STATUS" default:"88"` // 0x58 throttling error
	NoResponsePercent float64 `envconfig:"MNOSIM_NO_RESPONSE_PERCENT" default:"0"`
}

// LoadConfig reads the configuration and checks it.
func LoadConfig() (Config, error) {
	var c Config
	if err := envconfig.Process("", &c); err != nil {
		return c, err
	}
	if _, err := parseMix(c.DLRMix); err != nil {
		return c, err
	}
	if c.DLRDelayMin < 0 || c.DLRDelayMax < c.DLRDelayMin {
		return c, fmt.Errorf("MNOSIM_DLR_DELAY_MAX must be at least MNOSIM_DLR_DELAY_MIN")
	}
	if c.RejectPercent < 0 || c.NoResponsePercent < 0 || c.RejectPercent+c.NoResponsePercent > 100 {
		return c, fmt.Errorf("MNOSIM_REJECT_PERCENT and MNOSIM_NO_RESPONSE_PERCENT must be 0 to 100 and add up to at most 100")
	}
	return c, nil
}

type outcome struct {
	status string // DLR stat value, or "" for no DLR
	weight float64
}

var validStatuses = map[string]bool{"DELIVRD": true, "UNDELIV": true, "EXPIRED": true, "REJECTD": true, "DELETED": true, "UNKNOWN": true, "NONE": true}

func parseMix(s string) ([]outcome, error) {
	var out []outcome
	total := 0.0
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		name, w, ok := strings.Cut(part, ":")
		name = strings.ToUpper(strings.TrimSpace(name))
		weight, err := strconv.ParseFloat(strings.TrimSpace(w), 64)
		if !ok || err != nil || weight < 0 || !validStatuses[name] {
			return nil, fmt.Errorf("invalid MNOSIM_DLR_MIX entry %q (want STATUS:WEIGHT, STATUS one of DELIVRD, UNDELIV, EXPIRED, REJECTD, DELETED, UNKNOWN, NONE)", part)
		}
		if name == "NONE" {
			name = ""
		}
		out = append(out, outcome{status: name, weight: weight})
		total += weight
	}
	if total <= 0 {
		return nil, fmt.Errorf("MNOSIM_DLR_MIX must have a positive total weight")
	}
	return out, nil
}
