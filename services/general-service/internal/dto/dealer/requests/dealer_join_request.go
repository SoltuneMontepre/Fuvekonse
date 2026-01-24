package requests

type DealerJoinRequest struct {
	BoothCode string `json:"booth_code" binding:"required,len=6"`
}
