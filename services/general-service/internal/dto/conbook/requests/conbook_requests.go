package requests

// CreateConbookRequest is the request body for uploading a conbook
type CreateConbookRequest struct {
	Title       string `json:"title" binding:"required,min=1,max=255"`
	Description string `json:"description" binding:"max=500"`
	Handle      string `json:"handle" binding:"required,min=1,max=255"`
	ImageUrl    string `json:"image_url" binding:"required,url,max=500"`
}

// UpdateConbookRequest is the request body for updating a conbook (only before verification)
type UpdateConbookRequest struct {
	Title       *string `json:"title" binding:"omitempty,min=1,max=255"`
	Description *string `json:"description" binding:"omitempty,max=500"`
	Handle      *string `json:"handle" binding:"omitempty,min=1,max=255"`
	ImageUrl    *string `json:"image_url" binding:"omitempty,url,max=500"`
}

// VerifyConbookRequest is the request body for staff to verify a conbook
type VerifyConbookRequest struct {
	Approved bool   `json:"approved" binding:"required"`
	Reason   string `json:"reason" binding:"max=500"` // Reason for rejection if not approved
}
