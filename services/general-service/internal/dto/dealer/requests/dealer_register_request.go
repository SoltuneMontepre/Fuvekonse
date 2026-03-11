package requests

type DealerRegisterRequest struct {
	BoothName   string   `json:"booth_name" binding:"required,min=1,max=255"`
	Description string   `json:"description" binding:"required,min=1,max=500"`
	PriceSheets []string `json:"price_sheets" binding:"required,min=1,dive,required,url,max=500"`
}
