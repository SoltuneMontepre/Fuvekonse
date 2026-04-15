package requests

// CreateConbookRequest is the request body for uploading a conbook
type CreateConbookRequest struct {
	Title       string `json:"title" binding:"required,min=1,max=255"`
	Description string `json:"description" binding:"max=500"`
	Handle      string `json:"handle" binding:"required,min=1,max=255"`
	ImageUrl    string `json:"image_url" binding:"required,url,max=500"`
}

// UpdateConbookRequest is the request body for updating a conbook (only while pending)
type UpdateConbookRequest struct {
	Title       *string `json:"title" binding:"omitempty,min=1,max=255"`
	Description *string `json:"description" binding:"omitempty,max=500"`
	Handle      *string `json:"handle" binding:"omitempty,min=1,max=255"`
	ImageUrl    *string `json:"image_url" binding:"omitempty,url,max=500"`
}

// UpdateConbookStatusRequest is the request body for staff/admin status updates.
type UpdateConbookStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=pending approved denied"`
}
