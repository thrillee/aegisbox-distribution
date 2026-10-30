package sms

import (
	"context"
	"testing"

	"github.com/thrillee/aegisbox/internal/database"
	"github.com/thrillee/aegisbox/internal/mno"
)

// lateDLRQuerier returns one segment and records whether its DLR was overwritten.
type lateDLRQuerier struct {
	database.Querier
	segment database.FindSegmentByMnoMessageIDRow
	updated bool
}

func (q *lateDLRQuerier) FindSegmentByMnoMessageID(context.Context, *string) (database.FindSegmentByMnoMessageIDRow, error) {
	return q.segment, nil
}

func (q *lateDLRQuerier) UpdateSegmentDLR(context.Context, database.UpdateSegmentDLRParams) error {
	q.updated = true
	return nil
}

// A real DLR arriving after a manual or auto DLR completed the segment must not change it or be
// forwarded again to the SP.
func TestLateRealDLRIgnoredAfterManualOrAutoDLR(t *testing.T) {
	for _, source := range []string{DLRSourceAuto, DLRSourceManual} {
		status, src := "DELIVRD", source
		q := &lateDLRQuerier{segment: database.FindSegmentByMnoMessageIDRow{SegmentID: 1, MessageID: 2, DlrStatus: &status, DlrSource: &src}}
		svc := NewDLRService(DLRServiceDependencies{DBQueries: q})

		if err := svc.UpdateSegmentDLRStatus(context.Background(), 1, mno.DLRInfo{MnoMessageID: "ICN-1", Status: "UNDELIV"}); err != nil {
			t.Fatalf("%s: unexpected error: %v", source, err)
		}
		if q.updated {
			t.Fatalf("%s: a late real DLR overwrote a segment completed by a %s DLR", source, source)
		}
	}
}
