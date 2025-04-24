package segmenter

import (
	"errors"
	"fmt"
	"log/slog"
	"unicode/utf16"
)

const (
	// Max lengths per segment (approximate, depends on exact UDH)
	maxGSM7Single    = 160
	maxGSM7Multipart = 153 // 160 - 7 bytes for UDH
	maxUCS2Single    = 70
	maxUCS2Multipart = 67 // 70 - 3 code units (6 bytes) for UDH
)

// Segmenter defines the interface for splitting messages.
type Segmenter interface {
	// GetSegments splits a message, returning segments and indicating if UCS2 encoding is needed.
	GetSegments(message string) (segments []string, requiresUCS2 bool, err error)
}

// DefaultSegmenter provides a basic implementation.
type DefaultSegmenter struct{}

// NewDefaultSegmenter creates a basic segmenter.
func NewDefaultSegmenter() *DefaultSegmenter {
	return &DefaultSegmenter{}
}

// isGSM7 checks if a string contains only GSM-7 characters (simplified check).
// See https://en.wikipedia.org/wiki/GSM_03.38#GSM_7-bit_default_alphabet_and_extension_table_of_characters
func isGSM7(s string) bool {
	// This is a very basic check. A real implementation needs the full table including extensions.
	for _, r := range s {
		if r > 0x7F { // Basic ASCII range check, fails for extended chars
			// Check common extended chars? € is \x1b\x65, @ is \x00 etc.
			// For simplicity, assume anything outside basic ASCII requires UCS2.
			// A production system needs a proper GSM-7 encoder/validator.
			return false
		}
	}
	return true
}

// GetSegments implements the basic segmentation logic.
func (s *DefaultSegmenter) GetSegments(message string) ([]string, bool, error) {
	if message == "" {
		return []string{""}, false, nil // Empty message is one segment
	}

	requiresUCS2 := !isGSM7(message)
	var maxLength int
	var segments []string

	if requiresUCS2 {
		// Encode to UTF-16 code units for length calculation
		runes := utf16.Encode([]rune(message))
		messageLen := len(runes)

		if messageLen <= maxUCS2Single {
			maxLength = maxUCS2Single // Can fit in one UCS2 SMS
		} else {
			maxLength = maxUCS2Multipart
		}

		if maxLength <= 0 {
			return nil, true, errors.New("invalid UCS2 segment length configuration")
		}

		// Iterate through UTF-16 code units
		currentPos := 0
		for currentPos < messageLen {
			endPos := currentPos + maxLength
			if endPos > messageLen {
				endPos = messageLen
			}
			// Decode the segment back to UTF-8 string
			segmentRunes := utf16.Decode(runes[currentPos:endPos])
			segments = append(segments, string(segmentRunes))
			currentPos = endPos
		}
		slog.Debug("Segmented message using UCS2", slog.Int("segments", len(segments)), slog.Int("original_len_runes", messageLen))

	} else { // GSM-7
		messageLen := len(message) // GSM-7 chars (mostly) map 1:1 to bytes in simple cases

		if messageLen <= maxGSM7Single {
			maxLength = maxGSM7Single
		} else {
			maxLength = maxGSM7Multipart
		}

		if maxLength <= 0 {
			return nil, false, errors.New("invalid GSM7 segment length configuration")
		}

		currentPos := 0
		for currentPos < messageLen {
			endPos := currentPos + maxLength
			if endPos > messageLen {
				endPos = messageLen
			}
			segments = append(segments, message[currentPos:endPos])
			currentPos = endPos
		}
		slog.Debug("Segmented message using GSM7", slog.Int("segments", len(segments)), slog.Int("original_len_chars", messageLen))
	}

	if len(segments) == 0 && message != "" {
		// Should not happen with logic above, but safeguard
		return nil, requiresUCS2, fmt.Errorf("segmentation resulted in zero segments for non-empty message")
	}
	if len(segments) == 0 && message == "" {
		return []string{""}, false, nil // Return empty string for empty message
	}

	return segments, requiresUCS2, nil
}
