package requests

type DealerRegisterRequest struct {
	BoothName   string `json:"booth_name" binding:"required,min=1,max=255"`
	Description string `json:"description" binding:"required,min=1,max=500"`
	PriceSheet  string `json:"price_sheet" binding:"required,url,max=500"`
}
