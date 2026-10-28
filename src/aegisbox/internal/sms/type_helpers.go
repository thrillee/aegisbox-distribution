package sms

import (
	"fmt"
	"math/big"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
)

// Type conversion helper functions

// stringPtrToStr converts *string to string, returning empty string if nil
func stringPtrToStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// int32PtrToInt32 converts *int32 to int32, returning defaultVal if nil
func int32PtrToInt32(i *int32, defaultVal int32) int32 {
	if i == nil {
		return defaultVal
	}
	return *i
}

// numericToFloat64 converts pgtype.Numeric to float64
func numericToFloat64(n pgtype.Numeric) float64 {
	if !n.Valid {
		return 0.0
	}
	// Convert via string for accuracy
	f := new(big.Float).SetInt(n.Int)
	if n.Exp != 0 {
		// Apply the exponent (scale)
		scale := new(big.Float).SetFloat64(10.0)
		scale.SetMantExp(scale, int(n.Exp))
		f.Mul(f, scale)
	}
	result, _ := f.Float64()
	return result
}

// float64ToPgNumeric converts float64 to pgtype.Numeric
func float64ToPgNumeric(f float64) pgtype.Numeric {
	// Convert to string with 4 decimal places
	str := fmt.Sprintf("%.4f", f)
	bigF, _, _ := big.ParseFloat(str, 10, 256, big.ToNearestEven)

	// Scale to 4 decimal places (exp = -4)
	scaled := new(big.Float).Mul(bigF, big.NewFloat(10000))
	mantissa, _ := scaled.Int(nil)

	return pgtype.Numeric{
		Int:   mantissa,
		Exp:   -4,
		Valid: true,
	}
}

// decimalToFloat64 converts decimal.Decimal to float64
func decimalToFloat64(d decimal.Decimal) float64 {
	f, _ := d.Float64()
	return f
}
