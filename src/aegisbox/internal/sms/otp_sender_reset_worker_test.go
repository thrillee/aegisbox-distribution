package sms

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/thrillee/aegisbox/internal/database"
)

// resetWorkerMockQuerier is a minimal stub that only implements
// ResetDepletedOtpAlternativeSenders. All other methods panic to make
// unintended calls obvious immediately.
type resetWorkerMockQuerier struct {
	database.Querier // embed for compile-time satisfaction of the interface

	resetFunc func(ctx context.Context) ([]database.ResetDepletedOtpAlternativeSendersRow, error)
}

func (m *resetWorkerMockQuerier) ResetDepletedOtpAlternativeSenders(ctx context.Context) ([]database.ResetDepletedOtpAlternativeSendersRow, error) {
	return m.resetFunc(ctx)
}

// --- helpers ---

func makeRow(id int32, sid string) database.ResetDepletedOtpAlternativeSendersRow {
	return database.ResetDepletedOtpAlternativeSendersRow{ID: id, SenderIDString: sid}
}

// --- tests ---

// TestResetDepletedSenders_NoEligible verifies that when the DB returns an
// empty slice (no senders are due for reset), the worker returns 0, nil.
func TestResetDepletedSenders_NoEligible(t *testing.T) {
	mock := &resetWorkerMockQuerier{
		resetFunc: func(_ context.Context) ([]database.ResetDepletedOtpAlternativeSendersRow, error) {
			return []database.ResetDepletedOtpAlternativeSendersRow{}, nil
		},
	}

	worker := NewOtpSenderResetWorker(mock)
	count, err := worker.ResetDepletedSenders(context.Background(), 0)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected count=0, got: %d", count)
	}
}

// TestResetDepletedSenders_MultipleEligible verifies that when the DB returns
// multiple rows, the worker returns the correct count and no error.
func TestResetDepletedSenders_MultipleEligible(t *testing.T) {
	rows := []database.ResetDepletedOtpAlternativeSendersRow{
		makeRow(1, "AAI"),
		makeRow(2, "ABase"),
		makeRow(3, "ACloud"),
	}

	mock := &resetWorkerMockQuerier{
		resetFunc: func(_ context.Context) ([]database.ResetDepletedOtpAlternativeSendersRow, error) {
			return rows, nil
		},
	}

	worker := NewOtpSenderResetWorker(mock)
	count, err := worker.ResetDepletedSenders(context.Background(), 0)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if count != len(rows) {
		t.Fatalf("expected count=%d, got: %d", len(rows), count)
	}
}

// TestResetDepletedSenders_SingleEligible verifies the single-row happy path.
func TestResetDepletedSenders_SingleEligible(t *testing.T) {
	mock := &resetWorkerMockQuerier{
		resetFunc: func(_ context.Context) ([]database.ResetDepletedOtpAlternativeSendersRow, error) {
			return []database.ResetDepletedOtpAlternativeSendersRow{makeRow(42, "ZZone")}, nil
		},
	}

	worker := NewOtpSenderResetWorker(mock)
	count, err := worker.ResetDepletedSenders(context.Background(), 0)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected count=1, got: %d", count)
	}
}

// TestResetDepletedSenders_DBError verifies that a DB error is propagated
// as-is and count is 0.
func TestResetDepletedSenders_DBError(t *testing.T) {
	sentinel := errors.New("db connection refused")

	mock := &resetWorkerMockQuerier{
		resetFunc: func(_ context.Context) ([]database.ResetDepletedOtpAlternativeSendersRow, error) {
			return nil, sentinel
		},
	}

	worker := NewOtpSenderResetWorker(mock)
	count, err := worker.ResetDepletedSenders(context.Background(), 0)

	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected count=0 on error, got: %d", count)
	}
}

// TestResetDepletedSenders_ContextCancelled verifies that a cancelled context
// is propagated correctly from the DB layer.
func TestResetDepletedSenders_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	mock := &resetWorkerMockQuerier{
		resetFunc: func(c context.Context) ([]database.ResetDepletedOtpAlternativeSendersRow, error) {
			return nil, c.Err()
		},
	}

	worker := NewOtpSenderResetWorker(mock)
	count, err := worker.ResetDepletedSenders(ctx, 0)

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected count=0, got: %d", count)
	}
}

// TestResetDepletedSenders_BatchSizeIgnored verifies that the batchSize
// argument (unused — the SQL query handles it internally) has no effect
// on correctness regardless of the value passed.
func TestResetDepletedSenders_BatchSizeIgnored(t *testing.T) {
	rows := []database.ResetDepletedOtpAlternativeSendersRow{
		makeRow(10, "TOne"),
		makeRow(11, "TGo"),
	}

	mock := &resetWorkerMockQuerier{
		resetFunc: func(_ context.Context) ([]database.ResetDepletedOtpAlternativeSendersRow, error) {
			return rows, nil
		},
	}

	worker := NewOtpSenderResetWorker(mock)

	for _, batchSize := range []int{0, 1, 100, 9999} {
		count, err := worker.ResetDepletedSenders(context.Background(), batchSize)
		if err != nil {
			t.Fatalf("batchSize=%d: unexpected error: %v", batchSize, err)
		}
		if count != len(rows) {
			t.Fatalf("batchSize=%d: expected count=%d, got=%d", batchSize, len(rows), count)
		}
	}
}

// TestResetDepletedSenders_NilSliceFromDB verifies that a nil slice return
// (which pgx can return when no rows match) is handled safely.
func TestResetDepletedSenders_NilSliceFromDB(t *testing.T) {
	mock := &resetWorkerMockQuerier{
		resetFunc: func(_ context.Context) ([]database.ResetDepletedOtpAlternativeSendersRow, error) {
			return nil, nil // pgx returns nil slice when no rows updated
		},
	}

	worker := NewOtpSenderResetWorker(mock)
	count, err := worker.ResetDepletedSenders(context.Background(), 0)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected count=0, got: %d", count)
	}
}

// TestResetDepletedSenders_CallsDBExactlyOnce verifies the worker only makes
// one DB call per invocation, regardless of how many rows are returned.
func TestResetDepletedSenders_CallsDBExactlyOnce(t *testing.T) {
	callCount := 0

	mock := &resetWorkerMockQuerier{
		resetFunc: func(_ context.Context) ([]database.ResetDepletedOtpAlternativeSendersRow, error) {
			callCount++
			return []database.ResetDepletedOtpAlternativeSendersRow{
				makeRow(1, "AAI"),
				makeRow(2, "ABase"),
			}, nil
		},
	}

	worker := NewOtpSenderResetWorker(mock)
	_, _ = worker.ResetDepletedSenders(context.Background(), 0)

	if callCount != 1 {
		t.Fatalf("expected DB to be called exactly once, got: %d", callCount)
	}
}

// TestResetDepletedSenders_ReturnsCorrectIDs verifies that the returned count
// matches exactly the number of rows the DB reports as reset — no off-by-one.
func TestResetDepletedSenders_ReturnsCorrectIDs(t *testing.T) {
	cases := []struct {
		name     string
		rows     []database.ResetDepletedOtpAlternativeSendersRow
		wantCount int
	}{
		{"zero rows", nil, 0},
		{"one row", []database.ResetDepletedOtpAlternativeSendersRow{makeRow(1, "AAI")}, 1},
		{"five rows", []database.ResetDepletedOtpAlternativeSendersRow{
			makeRow(1, "AAI"), makeRow(2, "ABase"), makeRow(3, "ACloud"),
			makeRow(4, "ACore"), makeRow(5, "ADrive"),
		}, 5},
		{"391 rows (full Jennie set)", func() []database.ResetDepletedOtpAlternativeSendersRow {
			r := make([]database.ResetDepletedOtpAlternativeSendersRow, 391)
			for i := range r {
				r[i] = makeRow(int32(i+1), "SID")
			}
			return r
		}(), 391},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mock := &resetWorkerMockQuerier{
				resetFunc: func(_ context.Context) ([]database.ResetDepletedOtpAlternativeSendersRow, error) {
					return tc.rows, nil
				},
			}
			worker := NewOtpSenderResetWorker(mock)
			count, err := worker.ResetDepletedSenders(context.Background(), 0)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if count != tc.wantCount {
				t.Fatalf("expected %d, got %d", tc.wantCount, count)
			}
		})
	}
}

// TestResetDepletedSenders_WorkerFuncTimeout verifies that when a context
// deadline is exceeded mid-call, the error is surfaced correctly.
func TestResetDepletedSenders_WorkerFuncTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// Simulate slow DB by sleeping longer than the deadline
	mock := &resetWorkerMockQuerier{
		resetFunc: func(c context.Context) ([]database.ResetDepletedOtpAlternativeSendersRow, error) {
			select {
			case <-time.After(50 * time.Millisecond):
				return nil, nil
			case <-c.Done():
				return nil, c.Err()
			}
		},
	}

	worker := NewOtpSenderResetWorker(mock)
	count, err := worker.ResetDepletedSenders(ctx, 0)

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected DeadlineExceeded, got: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected count=0, got: %d", count)
	}
}

// --- compile-time stubs for the embedded Querier methods ---
// The embed satisfies the interface at compile time; these are only called
// if a test accidentally exercises a non-mocked method.

func (m *resetWorkerMockQuerier) AssignRoutingGroupToSpCredential(ctx context.Context, arg database.AssignRoutingGroupToSpCredentialParams) (database.SpCredential, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) CountMNOConnectionsByMNO(ctx context.Context, mnoID int32) (int64, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) CountMNOs(ctx context.Context) (int64, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) CountMessageDetails(ctx context.Context, arg database.CountMessageDetailsParams) (int64, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) CountMsisdnPrefixEntriesByGroup(ctx context.Context, msisdnPrefixGroupID int32) (int64, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) CountMsisdnPrefixGroups(ctx context.Context) (int64, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) CountOtpAlternativeSenders(ctx context.Context, arg database.CountOtpAlternativeSendersParams) (int64, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) CountOtpMessageTemplates(ctx context.Context, status *string) (int64, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) CountOtpSenderTemplateAssignments(ctx context.Context, arg database.CountOtpSenderTemplateAssignmentsParams) (int64, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) CountPricingRulesBySP(ctx context.Context, serviceProviderID int32) (int64, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) CountRoutingAssignmentsByRoutingGroup(ctx context.Context, routingGroupID int32) (int64, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) CountRoutingGroups(ctx context.Context) (int64, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) CountSPCredentials(ctx context.Context, dollar_1 int32) (int64, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) CountServiceProviders(ctx context.Context) (int64, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) CountWalletsBySP(ctx context.Context, serviceProviderID int32) (int64, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) CreateMNO(ctx context.Context, arg database.CreateMNOParams) (database.Mno, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) CreateMNOConnection(ctx context.Context, arg database.CreateMNOConnectionParams) (database.MnoConnection, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) CreateMessageSegment(ctx context.Context, arg database.CreateMessageSegmentParams) (int64, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) CreateMsisdnPrefixEntry(ctx context.Context, arg database.CreateMsisdnPrefixEntryParams) (database.MsisdnPrefixEntry, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) CreateMsisdnPrefixGroup(ctx context.Context, arg database.CreateMsisdnPrefixGroupParams) (database.MsisdnPrefixGroup, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) CreateOtpAlternativeSender(ctx context.Context, arg database.CreateOtpAlternativeSenderParams) (database.OtpAlternativeSender, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) CreateOtpMessageTemplate(ctx context.Context, arg database.CreateOtpMessageTemplateParams) (database.OtpMessageTemplate, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) CreateOtpSenderTemplateAssignment(ctx context.Context, arg database.CreateOtpSenderTemplateAssignmentParams) (database.OtpSenderTemplateAssignment, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) CreatePricingRule(ctx context.Context, arg database.CreatePricingRuleParams) (database.PricingRule, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) CreateRoutingAssignment(ctx context.Context, arg database.CreateRoutingAssignmentParams) (database.RoutingAssignment, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) CreateRoutingGroup(ctx context.Context, arg database.CreateRoutingGroupParams) (database.RoutingGroup, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) CreateSMPPCredential(ctx context.Context, arg database.CreateSMPPCredentialParams) (database.SpCredential, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) CreateSPCredential(ctx context.Context, arg database.CreateSPCredentialParams) (database.SpCredential, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) CreateServiceProvider(ctx context.Context, arg database.CreateServiceProviderParams) (database.ServiceProvider, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) CreateWallet(ctx context.Context, arg database.CreateWalletParams) (database.Wallet, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) CreateWalletTransaction(ctx context.Context, arg database.CreateWalletTransactionParams) (database.WalletTransaction, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) DeactivateMNO(ctx context.Context, id int32) (database.Mno, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) DeleteMNO(ctx context.Context, id int32) error {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) DeleteMNOConnection(ctx context.Context, id int32) error {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) DeleteMsisdnPrefixEntriesByGroup(ctx context.Context, msisdnPrefixGroupID int32) error {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) DeleteMsisdnPrefixEntry(ctx context.Context, id int32) error {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) DeleteMsisdnPrefixGroup(ctx context.Context, id int32) error {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) DeleteOtpAlternativeSender(ctx context.Context, id int32) error {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) DeleteOtpMessageTemplate(ctx context.Context, id int32) error {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) DeleteOtpSenderTemplateAssignment(ctx context.Context, id int32) error {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) DeletePricingRule(ctx context.Context, id int32) error {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) DeleteRoutingAssignment(ctx context.Context, id int32) error {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) DeleteRoutingGroup(ctx context.Context, id int32) error {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) DeleteSPCredential(ctx context.Context, id int32) error {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) DeleteServiceProvider(ctx context.Context, id int32) error {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) EnqueueDLRForForwarding(ctx context.Context, arg database.EnqueueDLRForForwardingParams) error {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) FindDebitTransactionForMessage(ctx context.Context, messageID *int64) (database.FindDebitTransactionForMessageRow, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) FindMatchingPrefixGroupForMSISDN(ctx context.Context, msisdnPrefix string) (database.FindMatchingPrefixGroupForMSISDNRow, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) FindMessageByMnoMessageID(ctx context.Context, mnoMessageID *string) (database.FindMessageByMnoMessageIDRow, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) FindSegmentByMnoMessageID(ctx context.Context, mnoMessageID *string) (database.FindSegmentByMnoMessageIDRow, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) GetActiveMNOConnections(ctx context.Context) ([]database.MnoConnection, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) GetActiveOtpTemplateAssignment(ctx context.Context, arg database.GetActiveOtpTemplateAssignmentParams) (database.GetActiveOtpTemplateAssignmentRow, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) GetApplicableMnoForRouting(ctx context.Context, arg database.GetApplicableMnoForRoutingParams) (database.GetApplicableMnoForRoutingRow, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) GetApplicablePrice(ctx context.Context, arg database.GetApplicablePriceParams) (database.GetApplicablePriceRow, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) GetDailyReport(ctx context.Context, arg database.GetDailyReportParams) ([]database.GetDailyReportRow, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) GetMNOByID(ctx context.Context, id int32) (database.Mno, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) GetMNOConnectionByID(ctx context.Context, id int32) (database.MnoConnection, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) GetMessageDetailsForPricing(ctx context.Context, id int64) (database.GetMessageDetailsForPricingRow, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) GetMessageDetailsForSending(ctx context.Context, id int64) (database.GetMessageDetailsForSendingRow, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) GetMessageFinalStatus(ctx context.Context, id int64) (string, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) GetMessageTotalSegments(ctx context.Context, id int64) (int32, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) GetMessagesByStatus(ctx context.Context, arg database.GetMessagesByStatusParams) ([]database.GetMessagesByStatusRow, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) GetMessagesReadyForRetry(ctx context.Context, limit int32) ([]database.GetMessagesReadyForRetryRow, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) GetMessagesToPrice(ctx context.Context, limit int32) ([]database.GetMessagesToPriceRow, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) GetMsisdnPrefixEntryByID(ctx context.Context, id int32) (database.MsisdnPrefixEntry, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) GetMsisdnPrefixEntryByPrefix(ctx context.Context, msisdnPrefix string) (database.MsisdnPrefixEntry, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) GetMsisdnPrefixGroupByID(ctx context.Context, id int32) (database.MsisdnPrefixGroup, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) GetMsisdnPrefixGroupByReference(ctx context.Context, reference *string) (database.MsisdnPrefixGroup, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) GetOtpAlternativeSenderByID(ctx context.Context, id int32) (database.GetOtpAlternativeSenderByIDRow, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) GetOtpMessageTemplateByID(ctx context.Context, id int32) (database.OtpMessageTemplate, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) GetOtpSenderTemplateAssignmentByID(ctx context.Context, id int32) (database.GetOtpSenderTemplateAssignmentByIDRow, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) GetPendingDLRsToForward(ctx context.Context, arg database.GetPendingDLRsToForwardParams) ([]database.GetPendingDLRsToForwardRow, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) GetPendingDLRsToForwardWithBackoff(ctx context.Context, arg database.GetPendingDLRsToForwardWithBackoffParams) ([]database.GetPendingDLRsToForwardWithBackoffRow, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) GetPricingRuleByID(ctx context.Context, id int32) (database.GetPricingRuleByIDRow, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) GetRoutingAssignmentByID(ctx context.Context, id int32) (database.GetRoutingAssignmentByIDRow, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) GetRoutingAssignmentByRoutingGroupAndPrefixGroup(ctx context.Context, arg database.GetRoutingAssignmentByRoutingGroupAndPrefixGroupParams) (database.RoutingAssignment, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) GetRoutingGroupByID(ctx context.Context, id int32) (database.RoutingGroup, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) GetRoutingGroupByReference(ctx context.Context, reference *string) (database.RoutingGroup, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) GetRoutingGroupIdForSpCredential(ctx context.Context, id int32) (database.GetRoutingGroupIdForSpCredentialRow, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) GetRoutingGroupReferenceById(ctx context.Context, id int32) (*string, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) GetSPCredentialByAPIKey(ctx context.Context, apiKeyHash *string) (database.GetSPCredentialByAPIKeyRow, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) GetSPCredentialByID(ctx context.Context, id int32) (database.SpCredential, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) GetSPCredentialByKeyIdentifier(ctx context.Context, apiKeyIdentifier *string) (database.GetSPCredentialByKeyIdentifierRow, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) GetSPCredentialBySystemID(ctx context.Context, systemID *string) (database.GetSPCredentialBySystemIDRow, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) GetSPMessageInfoForDLR(ctx context.Context, id int64) (database.GetSPMessageInfoForDLRRow, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) GetSegmentStatusesForMessage(ctx context.Context, messageID int64) ([]database.GetSegmentStatusesForMessageRow, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) GetServiceProviderByID(ctx context.Context, id int32) (database.ServiceProvider, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) GetSpCredentialRoutingGroupId(ctx context.Context, id int32) (*int32, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) GetSpCredentialScope(ctx context.Context, id int32) (string, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) GetTemplateContent(ctx context.Context, arg database.GetTemplateContentParams) (string, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) GetWalletByID(ctx context.Context, id int32) (database.Wallet, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) GetWalletForUpdate(ctx context.Context, arg database.GetWalletForUpdateParams) (database.GetWalletForUpdateRow, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) GetWalletForUpdateByID(ctx context.Context, id int32) (database.Wallet, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) GetWalletsBelowThreshold(ctx context.Context) ([]database.GetWalletsBelowThresholdRow, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) IncrementOtpAlternativeSenderUsage(ctx context.Context, id int32) (database.OtpAlternativeSender, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) IncrementOtpAssignmentUsage(ctx context.Context, id int32) (database.IncrementOtpAssignmentUsageRow, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) InsertMessageIn(ctx context.Context, arg database.InsertMessageInParams) (int64, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) ListAlternativeSendersWithAssignedTemplates(ctx context.Context, arg database.ListAlternativeSendersWithAssignedTemplatesParams) ([]database.ListAlternativeSendersWithAssignedTemplatesRow, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) ListGlobalOtpAlternativeSenders(ctx context.Context, arg database.ListGlobalOtpAlternativeSendersParams) ([]database.OtpAlternativeSender, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) ListMNOConnectionsByMNO(ctx context.Context, arg database.ListMNOConnectionsByMNOParams) ([]database.MnoConnection, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) ListMNOs(ctx context.Context, arg database.ListMNOsParams) ([]database.Mno, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) ListMessageDetails(ctx context.Context, arg database.ListMessageDetailsParams) ([]database.ListMessageDetailsRow, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) ListMsisdnPrefixEntriesByGroup(ctx context.Context, arg database.ListMsisdnPrefixEntriesByGroupParams) ([]database.MsisdnPrefixEntry, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) ListMsisdnPrefixGroups(ctx context.Context, arg database.ListMsisdnPrefixGroupsParams) ([]database.MsisdnPrefixGroup, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) ListOtpAlternativeSenders(ctx context.Context, arg database.ListOtpAlternativeSendersParams) ([]database.ListOtpAlternativeSendersRow, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) ListOtpMessageTemplates(ctx context.Context, arg database.ListOtpMessageTemplatesParams) ([]database.OtpMessageTemplate, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) ListOtpSenderTemplateAssignments(ctx context.Context, arg database.ListOtpSenderTemplateAssignmentsParams) ([]database.ListOtpSenderTemplateAssignmentsRow, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) ListPricingRulesBySP(ctx context.Context, arg database.ListPricingRulesBySPParams) ([]database.ListPricingRulesBySPRow, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) ListRoutingAssignmentsByRoutingGroup(ctx context.Context, arg database.ListRoutingAssignmentsByRoutingGroupParams) ([]database.ListRoutingAssignmentsByRoutingGroupRow, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) ListRoutingGroups(ctx context.Context, arg database.ListRoutingGroupsParams) ([]database.RoutingGroup, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) ListSPCredentials(ctx context.Context, arg database.ListSPCredentialsParams) ([]database.SpCredential, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) ListServiceProviders(ctx context.Context, arg database.ListServiceProvidersParams) ([]database.ServiceProvider, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) ListSpSpecificOtpAlternativeSenders(ctx context.Context, arg database.ListSpSpecificOtpAlternativeSendersParams) ([]database.OtpAlternativeSender, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) ListWalletsBySP(ctx context.Context, arg database.ListWalletsBySPParams) ([]database.Wallet, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) MarkDLRForwardingAttemptFailed(ctx context.Context, arg database.MarkDLRForwardingAttemptFailedParams) error {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) MarkDLRForwardingSuccess(ctx context.Context, id int64) error {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) MarkMessageForRetry(ctx context.Context, arg database.MarkMessageForRetryParams) error {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) MarkMessageSendFailed(ctx context.Context, arg database.MarkMessageSendFailedParams) error {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) MarkMessageSent(ctx context.Context, id int64) error {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) MoveToDeadLetterQueue(ctx context.Context, id int64) error {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) RemoveRoutingGroupFromSpCredential(ctx context.Context, spCredentialID int32) (database.SpCredential, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) UnlockStaleDLRs(ctx context.Context) error { panic("not implemented") }
func (m *resetWorkerMockQuerier) UpdateLowBalanceNotifiedAt(ctx context.Context, id int32) error {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) UpdateMNO(ctx context.Context, arg database.UpdateMNOParams) (database.Mno, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) UpdateMNOConnection(ctx context.Context, arg database.UpdateMNOConnectionParams) (database.MnoConnection, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) UpdateMessageDLRStatus(ctx context.Context, arg database.UpdateMessageDLRStatusParams) error {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) UpdateMessageFinalStatus(ctx context.Context, arg database.UpdateMessageFinalStatusParams) error {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) UpdateMessagePriced(ctx context.Context, arg database.UpdateMessagePricedParams) error {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) UpdateMessagePricedWithStatusTx(ctx context.Context, arg database.UpdateMessagePricedWithStatusTxParams) error {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) UpdateMessageRoutingInfo(ctx context.Context, arg database.UpdateMessageRoutingInfoParams) error {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) UpdateMessageSendAttempted(ctx context.Context, id int64) error {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) UpdateMessageValidatedRouted(ctx context.Context, arg database.UpdateMessageValidatedRoutedParams) error {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) UpdateMsisdnPrefixEntry(ctx context.Context, arg database.UpdateMsisdnPrefixEntryParams) (database.MsisdnPrefixEntry, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) UpdateMsisdnPrefixGroup(ctx context.Context, arg database.UpdateMsisdnPrefixGroupParams) (database.MsisdnPrefixGroup, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) UpdateOtpAlternativeSender(ctx context.Context, arg database.UpdateOtpAlternativeSenderParams) (database.OtpAlternativeSender, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) UpdateOtpMessageTemplate(ctx context.Context, arg database.UpdateOtpMessageTemplateParams) (database.OtpMessageTemplate, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) UpdateOtpSenderTemplateAssignment(ctx context.Context, arg database.UpdateOtpSenderTemplateAssignmentParams) (database.OtpSenderTemplateAssignment, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) UpdateRoutingAssignment(ctx context.Context, arg database.UpdateRoutingAssignmentParams) (database.RoutingAssignment, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) UpdateRoutingGroup(ctx context.Context, arg database.UpdateRoutingGroupParams) (database.RoutingGroup, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) UpdateSPCredential(ctx context.Context, arg database.UpdateSPCredentialParams) (database.SpCredential, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) UpdateSegmentDLR(ctx context.Context, arg database.UpdateSegmentDLRParams) error {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) UpdateSegmentSendFailed(ctx context.Context, arg database.UpdateSegmentSendFailedParams) error {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) UpdateSegmentSent(ctx context.Context, arg database.UpdateSegmentSentParams) error {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) UpdateServiceProvider(ctx context.Context, arg database.UpdateServiceProviderParams) (database.ServiceProvider, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) UpdateWalletBalance(ctx context.Context, arg database.UpdateWalletBalanceParams) (database.Wallet, error) {
	panic("not implemented")
}
func (m *resetWorkerMockQuerier) ValidateSenderID(ctx context.Context, arg database.ValidateSenderIDParams) (int32, error) {
	panic("not implemented")
}
