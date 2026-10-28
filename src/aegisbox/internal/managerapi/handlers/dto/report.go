package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

type DailyReportRow struct {
	ReportDate             time.Time       `json:"report_date"` // Should be just date, format on output
	ServiceProviderID      int32           `json:"service_provider_id"`
	ServiceProviderName    string          `json:"service_provider_name"`
	ServiceProviderEmail   string          `json:"service_provider_email"`
	MnoID                  int32           `json:"mno_id"`
	MnoName                string          `json:"mno_name"`
	CurrencyCode           string          `json:"currency_code"`
	TotalReceived          int64           `json:"total_received"` // Count of messages
	TotalSegmentsReceived  int64           `json:"total_segments_received"`
	TotalSegmentsSubmitted int64           `json:"total_segments_submitted"`
	TotalSegmentsDelivered int64           `json:"total_segments_delivered"`
	TotalSegmentsFailed    int64           `json:"total_segments_failed"`
	TotalSegmentsRejected  int64           `json:"total_segments_rejected"`
	TotalCostDebit         decimal.Decimal `json:"total_cost_debit"`    // Use decimal
	TotalCostReversal      decimal.Decimal `json:"total_cost_reversal"` // Use decimal
	NetCost                decimal.Decimal `json:"net_cost"`            // Calculated: Debit - Reversal
}
