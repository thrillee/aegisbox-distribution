package dto

import "time"

type RoutingAssignmentResponse struct {
	ID                    int32     `json:"id"`
	RoutingGroupID        int32     `json:"routing_group_id"`
	RoutingGroupName      *string   `json:"routing_group_name,omitempty"`
	RoutingGroupReference *string   `json:"routing_group_reference,omitempty"`
	MsisdnPrefixGroupID   int32     `json:"msisdn_prefix_group_id"`
	PrefixGroupName       *string   `json:"prefix_group_name,omitempty"`
	PrefixGroupReference  *string   `json:"prefix_group_reference,omitempty"`
	MnoID                 int32     `json:"mno_id"`
	MnoName               *string   `json:"mno_name,omitempty"`
	Status                string    `json:"status"`
	Priority              int32     `json:"priority"`
	Comment               *string   `json:"comment,omitempty"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

type CreateRoutingAssignmentRequest struct {
	RoutingGroupID      int32   `json:"routing_group_id" binding:"required"`
	MsisdnPrefixGroupID int32   `json:"msisdn_prefix_group_id" binding:"required"`
	MnoID               int32   `json:"mno_id" binding:"required"`
	Status              *string `json:"status" binding:"omitempty,oneof=active inactive"`
	Priority            *int32  `json:"priority"`
	Comment             *string `json:"comment"`
}

type UpdateRoutingAssignmentRequest struct {
	MnoID    *int32  `json:"mno_id"`
	Status   *string `json:"status" binding:"omitempty,oneof=active inactive"`
	Priority *int32  `json:"priority"`
	Comment  *string `json:"comment"`
}

type ListRoutingAssignmentsRequest struct {
	RoutingGroupID int32 `json:"routing_group_id" binding:"required"`
	Limit          int32 `json:"limit"`
	Offset         int32 `json:"offset"`
}
