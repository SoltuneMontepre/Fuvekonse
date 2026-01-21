package requests

type DealerRemoveStaffRequest struct {
	StaffUserId string `json:"staff_user_id" binding:"required,uuid"`
}
