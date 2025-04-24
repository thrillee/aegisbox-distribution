package errormapper

import (
	"log/slog"
	"strings"
)

// Basic map for internal codes to protocol-specific codes. Expand significantly for production.
var internalToSMPP = map[string]string{
	"NO_ROUTE":        "00B", // Invalid Destination Address
	"INVALID_SENDER":  "00A", // Invalid Source Address
	"INSUF_FUNDS":     "058", // ESME Receiver Temporary App Error Code (no specific funds error)
	"SYS_ERR":         "008", // System Error
	"MNO_REJECT":      "00D", // Submit Failed (generic MNO rejection)
	"VALIDATION_FAIL": "045", // Invalid Message ID (abusing slightly for generic validation fail)
	"DELIVRD":         "000", // Delivered (Not an error, but useful for mapping)
	// ... add mappings for common MNO error codes if known ...
	// e.g., "MNO_01": "00C" // Invalid Message Length from MNO
}

var internalToHTTP = map[string]string{
	"NO_ROUTE":        "400 Bad Request - No Route", // Example HTTP error string
	"INVALID_SENDER":  "400 Bad Request - Invalid Sender",
	"INSUF_FUNDS":     "402 Payment Required", // Fits semantically
	"SYS_ERR":         "500 Internal Server Error",
	"MNO_REJECT":      "502 Bad Gateway - MNO Rejected",
	"VALIDATION_FAIL": "400 Bad Request - Validation Failed",
	"DELIVRD":         "200 OK - Delivered",
	// ...
}

// MapErrorCode translates internal/MNO error codes to protocol-specific ones.
func MapErrorCode(internalCode string, protocol string) string {
	protocol = strings.ToLower(protocol)
	internalCode = strings.ToUpper(internalCode) // Normalize internal code

	var targetMap map[string]string
	var defaultCode string

	switch protocol {
	case "smpp":
		targetMap = internalToSMPP
		defaultCode = "008" // Default to System Error for SMPP
	case "http":
		targetMap = internalToHTTP
		defaultCode = "500 Internal Server Error" // Default to 500 for HTTP
	default:
		slog.Warn("Unknown protocol requested for error code mapping", slog.String("protocol", protocol))
		return internalCode // Return original code if protocol unknown
	}

	if mappedCode, ok := targetMap[internalCode]; ok {
		return mappedCode
	}

	slog.Debug("No specific mapping found for error code, returning default",
		slog.String("internal_code", internalCode),
		slog.String("protocol", protocol),
		slog.String("default_code", defaultCode),
	)
	return defaultCode
}
