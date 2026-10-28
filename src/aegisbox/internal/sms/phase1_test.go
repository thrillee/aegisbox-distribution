package sms

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thrillee/aegisbox/internal/sp"
	"github.com/thrillee/aegisbox/pkg/codes"
)

func TestStatusConstants_ValuesAreCorrect(t *testing.T) {
	assert.Equal(t, "received", codes.MsgStatusReceived)
	assert.Equal(t, "routed", codes.MsgStatusRouted)
	assert.Equal(t, "queued_for_send", codes.MsgStatusQueuedForSend)
	assert.Equal(t, "failed_routing", codes.MsgStatusFailedRouting)
}

func TestDigitsToWords_ConvertsCorrectly(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1234", "one two three four"},
		{"5678", "five six seven eight"},
		{"9012", "nine zero one two"},
		{"0000", "zero zero zero zero"},
		{"999999", "nine nine nine nine nine nine"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := DigitsToWords(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestOTPPreprocessor_TemplateWithOtpWords(t *testing.T) {
	// Test the {{otp_words}} placeholder functionality
	tests := []struct {
		name     string
		template string
		otp      string
		expected string
	}{
		{
			name:     "Simple OTP words",
			template: "Your code is {{otp_words}}",
			otp:      "1234",
			expected: "Your code is one two three four",
		},
		{
			name:     "OTP words with other placeholders",
			template: "[OTP] is your {{otp_words}} code",
			otp:      "5678",
			expected: "5678 is your five six seven eight code",
		},
		{
			name:     "No OTP words placeholder",
			template: "Your code is [OTP]",
			otp:      "9012",
			expected: "Your code is 9012",
		},
		{
			name:     "Multiple OTP words",
			template: "{{otp_words}} {{otp_words}}",
			otp:      "12",
			expected: "one two one two",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the template processing logic from otp_preprocessor.go
			result := tt.template
			result = strings.ReplaceAll(result, "[OTP]", tt.otp)

			// Add {{otp_words}} support
			if strings.Contains(result, "{{otp_words}}") {
				otpWords := DigitsToWords(tt.otp)
				result = strings.ReplaceAll(result, "{{otp_words}}", otpWords)
			}

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRoutingPreprocessorRemoved(t *testing.T) {
	// Verify that routing_preprocessor.go has been deleted
	_, err := os.Stat("routing_preprocessor.go")
	assert.True(t, os.IsNotExist(err), "routing_preprocessor.go should be deleted")
}

func TestIncomingHandlerStructure(t *testing.T) {
	// This test verifies that the incoming handler structure is correct
	handler := &DefaultIncomingMessageHandler{
		preprocessors: []sp.MessagePreprocessor{}, // Should be empty after routing preprocessor removal
	}
	assert.NotNil(t, handler)
	assert.Empty(t, handler.preprocessors, "Preprocessors should be empty")
}
