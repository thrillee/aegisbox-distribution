package sms

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

var (
	otpPatterns = []*regexp.Regexp{
		regexp.MustCompile(`\b(\d{4,6})\b`),
		regexp.MustCompile(`\b(\d{3}\s?\d{3})\b`),
		regexp.MustCompile(`\b(\d{4}-\d{4})\b`),
		regexp.MustCompile(`(?i)(?:code|otp|pin)[:\s]*(\d{4,6})`),
	}
	otpKeywords = regexp.MustCompile(`(?i)(otp|password|verification|code.*\d|pin.*\d|confirm.*\d)`)
)

var digitToWord = map[rune]string{
	'0': "zero",
	'1': "one",
	'2': "two",
	'3': "three",
	'4': "four",
	'5': "five",
	'6': "six",
	'7': "seven",
	'8': "eight",
	'9': "nine",
}

type OTPTemplate struct {
	Template        string
	BrandName       string
	AppName         string
	ValidityMinutes int
}

func IsOTPMessage(message string) bool {
	return otpKeywords.MatchString(message)
}

func ExtractOTP(message string) string {
	for _, pattern := range otpPatterns {
		matches := pattern.FindStringSubmatch(message)
		if len(matches) > 1 {
			return matches[1]
		}
	}
	return ""
}

func ApplyOTPTemplate(originalMessage, template string) string {
	otpDigits := ExtractOTP(originalMessage)
	formattedMessage := fmt.Sprintf(template, otpDigits)
	return formattedMessage
}

func ApplyOTPTemplateWithContext(ctx context.Context, originalMessage string, template OTPTemplate) string {
	otpDigits := ExtractOTP(originalMessage)
	if otpDigits == "" {
		return originalMessage
	}

	formattedMessage := template.Template
	formattedMessage = strings.ReplaceAll(formattedMessage, "[OTP]", otpDigits)
	formattedMessage = strings.ReplaceAll(formattedMessage, "[BRAND_NAME]", template.BrandName)
	formattedMessage = strings.ReplaceAll(formattedMessage, "[APP_NAME]", template.AppName)
	formattedMessage = strings.ReplaceAll(formattedMessage, "[VALIDITY_MINUTES]", itoa(template.ValidityMinutes))

	return formattedMessage
}

func DigitsToWords(s string) string {
	var result strings.Builder
	for _, r := range s {
		if word, ok := digitToWord[r]; ok {
			result.WriteString(word)
			result.WriteString(" ")
		} else {
			result.WriteRune(r)
		}
	}
	return strings.TrimSpace(result.String())
}

func ConvertMessageDigitsToWords(message string) string {
	pattern := regexp.MustCompile(`\d`)
	return pattern.ReplaceAllStringFunc(message, func(match string) string {
		return DigitsToWords(match)
	})
}

func GetOTPTemplateForSender(senderID string, templates map[string]OTPTemplate) (OTPTemplate, bool) {
	template, exists := templates[senderID]
	return template, exists
}

func BuildOTPTemplate(contentTemplate, defaultBrandName string) OTPTemplate {
	return OTPTemplate{
		Template:        contentTemplate,
		BrandName:       defaultBrandName,
		AppName:         "Aegisbox",
		ValidityMinutes: 5,
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
