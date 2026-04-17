package handlers

import (
	"general-service/internal/common/constants"
	"general-service/internal/common/utils"
	"general-service/internal/services"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

type DevMailHandler struct {
	services *services.Services
}

func NewDevMailHandler(services *services.Services) *DevMailHandler {
	return &DevMailHandler{services: services}
}

type devMailSendRequest struct {
	To            string `json:"to" binding:"required,email"`
	Kind          string `json:"kind" binding:"required"` // otp | dealer | ticket | ticket_denied
	Lang          string `json:"lang"`
	Otp           string `json:"otp"`
	BoothName     string `json:"boothName"`
	BoothNumber   string `json:"boothNumber"`
	ReferenceCode string `json:"referenceCode"`
	TierName      string `json:"tierName"`
	Reason        string `json:"reason"`
}

// SendTestMail triggers a single mail send for local/staging verification. Only registered when ENV != production.
func (h *DevMailHandler) SendTestMail(c *gin.Context) {
	var req devMailSendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondErrorWithErrorMessage(c, http.StatusBadRequest, constants.ErrCodeValidationFailed, err.Error(), "validationFailed")
		return
	}

	fromEmail := os.Getenv("SES_EMAIL_IDENTITY")
	if fromEmail == "" {
		utils.RespondErrorWithErrorMessage(c, http.StatusBadRequest, constants.ErrCodeBadRequest, "SES_EMAIL_IDENTITY is not set", "mailFromNotConfigured")
		return
	}

	lang := strings.TrimSpace(strings.ToLower(req.Lang))
	if lang != "vi" {
		lang = "en"
	}

	ctx := c.Request.Context()
	kind := strings.ToLower(strings.TrimSpace(req.Kind))

	var err error
	switch kind {
	case "otp":
		otp := strings.TrimSpace(req.Otp)
		if otp == "" {
			otp = "123456"
		}
		err = h.services.Mail.SendOtpEmail(ctx, fromEmail, req.To, otp, lang)
	case "dealer":
		boothName := strings.TrimSpace(req.BoothName)
		if boothName == "" {
			boothName = "Dev test booth"
		}
		boothNumber := strings.TrimSpace(req.BoothNumber)
		if boothNumber == "" {
			boothNumber = "B-DEV-01"
		}
		err = h.services.Mail.SendDealerApprovedEmail(ctx, fromEmail, req.To, boothName, boothNumber, lang)
	case "ticket":
		ref := strings.TrimSpace(req.ReferenceCode)
		if ref == "" {
			ref = "DEV-TICKET-REF"
		}
		tier := strings.TrimSpace(req.TierName)
		err = h.services.Mail.SendTicketApprovedWithQREmail(ctx, fromEmail, req.To, ref, tier, lang)
	case "ticket_denied":
		ref := strings.TrimSpace(req.ReferenceCode)
		if ref == "" {
			ref = "DEV-TICKET-REF"
		}
		tier := strings.TrimSpace(req.TierName)
		err = h.services.Mail.SendTicketDeniedEmail(ctx, fromEmail, req.To, ref, tier, strings.TrimSpace(req.Reason), lang)
	default:
		utils.RespondErrorWithErrorMessage(c, http.StatusBadRequest, constants.ErrCodeBadRequest, "kind must be otp, dealer, ticket, or ticket_denied", "invalidMailKind")
		return
	}

	if err != nil {
		utils.RespondErrorWithErrorMessage(c, http.StatusInternalServerError, constants.ErrCodeInternalServerError, err.Error(), "mailSendFailed")
		return
	}

	type devMailSendResponse struct {
		Kind string `json:"kind"`
		To   string `json:"to"`
	}
	utils.RespondSuccess(c, &devMailSendResponse{Kind: kind, To: req.To}, "mail sent")
}
