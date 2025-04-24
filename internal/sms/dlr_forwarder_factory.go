package sms

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/thrillee/aegisbox/internal/sp"
)

// compile-time check to ensure implementation satisfies the interface
var _ DLRForwarderFactory = (*mapDLRForwarderFactory)(nil)

// mapDLRForwarderFactory holds instances of concrete DLRForwarder implementations.
type mapDLRForwarderFactory struct {
	smppForwarder *sp.SMPSPForwarder  // Instance of the SMPP forwarder
	httpForwarder *sp.HTTPSPForwarder // Instance of the HTTP forwarder
	// Add other forwarders here if supporting more protocols (e.g., SMTP)
}

// NewMapDLRForwarderFactory creates a new factory instance.
// It requires pre-initialized instances of the concrete forwarders.
func NewMapDLRForwarderFactory(smppFwd *sp.SMPSPForwarder, httpFwd *sp.HTTPSPForwarder) (DLRForwarderFactory, error) {
	if smppFwd == nil {
		slog.Warn("SMPP DLR Forwarder instance is nil during factory creation. SMPP forwarding will not work.")
		// Depending on requirements, you might return an error here:
		// return nil, errors.New("SMPSPForwarder instance cannot be nil")
	}
	if httpFwd == nil {
		slog.Warn("HTTP DLR Forwarder instance is nil during factory creation. HTTP forwarding will not work.")
		// Depending on requirements, you might return an error here:
		// return nil, errors.New("HTTPSPForwarder instance cannot be nil")
	}

	return &mapDLRForwarderFactory{
		smppForwarder: smppFwd,
		httpForwarder: httpFwd,
	}, nil
}

// GetForwarder returns the appropriate DLRForwarder based on the protocol string.
func (f *mapDLRForwarderFactory) GetForwarder(protocol string) (sp.DLRForwarder, error) {
	normalizedProtocol := strings.ToLower(strings.TrimSpace(protocol))

	switch normalizedProtocol {
	case "smpp":
		if f.smppForwarder == nil {
			return nil, fmt.Errorf("SMPP forwarder requested but not initialized in factory")
		}
		slog.Debug("Providing SMPP DLR Forwarder instance")
		return f.smppForwarder, nil
	case "http":
		if f.httpForwarder == nil {
			return nil, fmt.Errorf("HTTP forwarder requested but not initialized in factory")
		}
		slog.Debug("Providing HTTP DLR Forwarder instance")
		return f.httpForwarder, nil
	// Add cases for other protocols if needed
	// case "smtp":
	//     // return f.smtpForwarder, nil
	default:
		slog.Warn("Request for unknown DLR forwarder protocol", slog.String("protocol", protocol))
		return nil, fmt.Errorf("unknown or unsupported DLR forwarder protocol: %s", protocol)
	}
}
