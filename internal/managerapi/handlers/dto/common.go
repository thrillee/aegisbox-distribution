package dto

// PaginationResponse provides standard pagination details.
type PaginationResponse struct {
	Total  int64 `json:"total"`  // Total number of records available
	Limit  int32 `json:"limit"`  // Number of records per page
	Offset int32 `json:"offset"` // Starting record index
}

// PaginatedListResponse is a generic wrapper for list API responses.
type PaginatedListResponse struct {
	Data       any                `json:"data"` // The actual list of DTOs
	Pagination PaginationResponse `json:"pagination"`
}
