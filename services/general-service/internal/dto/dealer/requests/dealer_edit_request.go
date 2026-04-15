package requests

type DealerEditRequest struct {
	BoothName   *string   `json:"booth_name" binding:"omitempty,min=1,max=255"`
	Description *string   `json:"description" binding:"omitempty,min=1,max=500"`
	PriceSheets *[]string `json:"price_sheets" binding:"omitempty,min=1,max=10,dive,required,url,max=500"`
}
