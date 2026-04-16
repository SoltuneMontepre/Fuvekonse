package services

import (
	"bytes"
	"context"
	"embed"
	"encoding/base64"
	"fmt"
	"general-service/internal/repositories"
	htemplate "html/template"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/skip2/go-qrcode"
)

const (
	mailProviderSES      = "ses"
	mailProviderSendGrid = "sendgrid"
)

//go:embed html/*.html
var mailHTML embed.FS

var mailTemplates *htemplate.Template

func init() {
	var err error
	mailTemplates, err = htemplate.New("").ParseFS(mailHTML, "html/*.html")
	if err != nil {
		log.Fatalf("parse mail HTML templates: %v", err)
	}
}

func renderMailTemplate(name string, data any) (string, error) {
	var buf bytes.Buffer
	if err := mailTemplates.ExecuteTemplate(&buf, name, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// LangFromCountry returns the email language code from the user's country (e.g. "vi" for Vietnam, "en" otherwise).
func LangFromCountry(country string) string {
	c := strings.TrimSpace(strings.ToLower(country))
	if c == "" {
		return "en"
	}
	switch c {
	case "vietnam", "vn", "việt nam", "viet nam":
		return "vi"
	default:
		if strings.Contains(c, "vietnam") || strings.Contains(c, "việt nam") {
			return "vi"
		}
		return "en"
	}
}

type MailService struct {
	repos           *repositories.Repositories
	sesClient       *ses.Client
	sendgridClient  *sendgrid.Client
	defaultFromName string
}

func NewMailService(repos *repositories.Repositories) *MailService {
	provider := strings.ToLower(strings.TrimSpace(os.Getenv("MAIL_PROVIDER")))
	if provider == "" {
		provider = mailProviderSES
	}

	var sesClient *ses.Client
	var sendgridClient *sendgrid.Client
	defaultFromName := os.Getenv("MAIL_FROM_NAME")
	if defaultFromName == "" {
		defaultFromName = "Fuvekon"
	}

	switch provider {
	case mailProviderSendGrid:
		apiKey := os.Getenv("SENDGRID_API_KEY")
		if apiKey == "" {
			log.Fatal("MAIL_PROVIDER=sendgrid requires SENDGRID_API_KEY to be set")
		}
		sendgridClient = sendgrid.NewSendClient(apiKey)
		log.Println("Mail service using SendGrid")
	default:
		sesClient = initSESClient()
		log.Println("Mail service using AWS SES")
	}

	return &MailService{
		repos:           repos,
		sesClient:       sesClient,
		sendgridClient:  sendgridClient,
		defaultFromName: defaultFromName,
	}
}

func initSESClient() *ses.Client {
	ctx := context.Background()
	region := os.Getenv("AWS_REGION")
	if region == "" {
		log.Fatal("AWS_REGION is not set (required when using SES)")
	}

	useLocalStack := os.Getenv("USE_LOCALSTACK") == "true"

	opts := []func(*config.LoadOptions) error{
		config.WithRegion(region),
	}

	if useLocalStack {
		localEndpoint := os.Getenv("LOCALSTACK_ENDPOINT")
		if localEndpoint == "" {
			log.Fatal("USE_LOCALSTACK=true but LOCALSTACK_ENDPOINT is not set")
		}
		accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
		secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
		if accessKey == "" || secretKey == "" {
			log.Fatal("USE_LOCALSTACK=true but AWS credentials are missing")
		}
		log.Println("Using LocalStack for SES")
		opts = append(opts,
			config.WithCredentialsProvider(
				credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
			),
			config.WithEndpointResolverWithOptions(
				aws.EndpointResolverWithOptionsFunc(
					func(service, region string, options ...interface{}) (aws.Endpoint, error) {
						return aws.Endpoint{
							URL:           localEndpoint,
							SigningRegion: region,
						}, nil
					},
				),
			),
		)
	}

	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		log.Fatalf("failed to load AWS SDK config: %v", err)
	}
	return ses.NewFromConfig(cfg)
}

func (s *MailService) SendEmail(
	ctx context.Context,
	fromEmail string,
	toEmail string,
	subject string,
	body string,
	cc []string,
	bcc []string,
) error {
	if s.sendgridClient != nil {
		return s.sendEmailSendGrid(ctx, fromEmail, toEmail, subject, body, cc, bcc)
	}
	if s.sesClient != nil {
		return s.sendEmailSES(ctx, fromEmail, toEmail, subject, body, cc, bcc)
	}
	return fmt.Errorf("no mail provider configured")
}

func (s *MailService) sendEmailSES(
	ctx context.Context,
	fromEmail, toEmail, subject, body string,
	cc, bcc []string,
) error {
	log.Printf("Sending email via SES (from=%s to=%s subject=%s)", fromEmail, toEmail, subject)

	destination := &types.Destination{
		ToAddresses: []string{toEmail},
	}
	if len(cc) > 0 {
		destination.CcAddresses = cc
	}
	if len(bcc) > 0 {
		destination.BccAddresses = bcc
	}

	input := &ses.SendEmailInput{
		Source:      aws.String(fromEmail),
		Destination: destination,
		Message: &types.Message{
			Subject: &types.Content{
				Data:    aws.String(subject),
				Charset: aws.String("UTF-8"),
			},
			Body: &types.Body{
				Html: &types.Content{
					Data:    aws.String(body),
					Charset: aws.String("UTF-8"),
				},
			},
		},
	}

	resp, err := s.sesClient.SendEmail(ctx, input)
	if err != nil {
		log.Printf("Failed to send email via SES (from=%s to=%s): %v", fromEmail, toEmail, err)
		return fmt.Errorf("SES SendEmail failed: %w", err)
	}
	log.Printf("Email sent via SES. MessageID=%s", aws.ToString(resp.MessageId))
	return nil
}

// sendEmailSESRawWithInlineImage sends an email via SES using SendRawEmail with a multipart/related body (HTML + inline image).
func (s *MailService) sendEmailSESRawWithInlineImage(ctx context.Context, fromEmail, toEmail, subject, htmlBody, contentID string, imagePNG []byte) error {
	log.Printf("Sending email via SES with inline image (from=%s to=%s subject=%s)", fromEmail, toEmail, subject)
	boundary := "ticket-qr-boundary"
	var buf bytes.Buffer
	buf.WriteString("From: " + fromEmail + "\r\n")
	buf.WriteString("To: " + toEmail + "\r\n")
	buf.WriteString("Subject: " + subject + "\r\n")
	buf.WriteString("MIME-Version: 1.0\r\n")
	buf.WriteString("Content-Type: multipart/related; boundary=\"" + boundary + "\"\r\n\r\n")

	// HTML part
	buf.WriteString("--" + boundary + "\r\n")
	buf.WriteString("Content-Type: text/html; charset=UTF-8\r\n\r\n")
	buf.WriteString(htmlBody)
	buf.WriteString("\r\n")

	// Inline image part
	buf.WriteString("--" + boundary + "\r\n")
	buf.WriteString("Content-Type: image/png\r\n")
	buf.WriteString("Content-Disposition: inline; filename=\"qrcode.png\"\r\n")
	buf.WriteString("Content-ID: <" + contentID + ">\r\n")
	buf.WriteString("Content-Transfer-Encoding: base64\r\n\r\n")
	b64 := base64.StdEncoding
	encoded := make([]byte, b64.EncodedLen(len(imagePNG)))
	b64.Encode(encoded, imagePNG)
	for i := 0; i < len(encoded); i += 76 {
		end := i + 76
		if end > len(encoded) {
			end = len(encoded)
		}
		buf.Write(encoded[i:end])
		buf.WriteString("\r\n")
	}
	buf.WriteString("--" + boundary + "--\r\n")

	input := &ses.SendRawEmailInput{
		Source:       aws.String(fromEmail),
		Destinations: []string{toEmail},
		RawMessage: &types.RawMessage{
			Data: buf.Bytes(),
		},
	}
	resp, err := s.sesClient.SendRawEmail(ctx, input)
	if err != nil {
		log.Printf("Failed to send raw email via SES (from=%s to=%s): %v", fromEmail, toEmail, err)
		return fmt.Errorf("SES SendRawEmail failed: %w", err)
	}
	log.Printf("Email sent via SES. MessageID=%s", aws.ToString(resp.MessageId))
	return nil
}

func (s *MailService) sendEmailSendGrid(
	ctx context.Context,
	fromEmail, toEmail, subject, body string,
	cc, bcc []string,
) error {
	log.Printf("Sending email via SendGrid (from=%s to=%s subject=%s)", fromEmail, toEmail, subject)

	from := mail.NewEmail(s.defaultFromName, fromEmail)
	to := mail.NewEmail("", toEmail)
	// SendGrid expects plain text and HTML; use body as HTML and a stripped version for plain
	plainText := stripHTML(body)
	if plainText == "" {
		plainText = subject
	}
	message := mail.NewSingleEmail(from, subject, to, plainText, body)

	if len(cc) > 0 && len(message.Personalizations) > 0 {
		ccEmails := make([]*mail.Email, 0, len(cc))
		for _, addr := range cc {
			ccEmails = append(ccEmails, mail.NewEmail("", addr))
		}
		message.Personalizations[0].AddCCs(ccEmails...)
	}
	if len(bcc) > 0 && len(message.Personalizations) > 0 {
		bccEmails := make([]*mail.Email, 0, len(bcc))
		for _, addr := range bcc {
			bccEmails = append(bccEmails, mail.NewEmail("", addr))
		}
		message.Personalizations[0].AddBCCs(bccEmails...)
	}

	resp, err := s.sendgridClient.Send(message)
	if err != nil {
		log.Printf("Failed to send email via SendGrid (from=%s to=%s): %v", fromEmail, toEmail, err)
		return fmt.Errorf("SendGrid Send failed: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Printf("SendGrid returned %d (from=%s to=%s): %s", resp.StatusCode, fromEmail, toEmail, resp.Body)
		return fmt.Errorf("SendGrid returned status %d: %s", resp.StatusCode, resp.Body)
	}
	log.Printf("Email sent via SendGrid successfully")
	return nil
}

// sendEmailSendGridWithInlineImage sends an email via SendGrid with an inline image (Content-ID).
func (s *MailService) sendEmailSendGridWithInlineImage(ctx context.Context, fromEmail, toEmail, subject, htmlBody, contentID string, imagePNG []byte) error {
	log.Printf("Sending email via SendGrid with inline image (from=%s to=%s subject=%s)", fromEmail, toEmail, subject)
	from := mail.NewEmail(s.defaultFromName, fromEmail)
	to := mail.NewEmail("", toEmail)
	plainText := stripHTML(htmlBody)
	if plainText == "" {
		plainText = subject
	}
	message := mail.NewSingleEmail(from, subject, to, plainText, htmlBody)
	attachment := mail.NewAttachment()
	attachment.SetContent(base64.StdEncoding.EncodeToString(imagePNG))
	attachment.SetType("image/png")
	attachment.SetFilename("qrcode.png")
	attachment.SetDisposition("inline")
	attachment.SetContentID(contentID)
	message.AddAttachment(attachment)

	resp, err := s.sendgridClient.Send(message)
	if err != nil {
		log.Printf("Failed to send email via SendGrid (from=%s to=%s): %v", fromEmail, toEmail, err)
		return fmt.Errorf("SendGrid Send failed: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Printf("SendGrid returned %d (from=%s to=%s): %s", resp.StatusCode, fromEmail, toEmail, resp.Body)
		return fmt.Errorf("SendGrid returned status %d: %s", resp.StatusCode, resp.Body)
	}
	log.Printf("Email sent via SendGrid successfully")
	return nil
}

// stripHTML removes HTML tags for plain-text fallback.
func stripHTML(s string) string {
	var b strings.Builder
	inTag := false
	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag && (r == ' ' || r == '\n' || r >= 0x20):
			b.WriteRune(r)
		}
	}
	return strings.TrimSpace(b.String())
}

// SendOtpEmail sends an OTP verification email to the specified email address. lang: "vi" for Vietnamese, else English.
func (s *MailService) SendOtpEmail(ctx context.Context, fromEmail, toEmail, otp, lang string) error {
	var subject, tpl string
	if lang == "vi" {
		subject = "Xác thực email của bạn cho FUVE"
		tpl = "otp_vi.html"
	} else {
		subject = "Verify your email for FUVE"
		tpl = "otp_en.html"
	}
	body, err := renderMailTemplate(tpl, struct{ Otp string }{Otp: otp})
	if err != nil {
		return fmt.Errorf("render otp email: %w", err)
	}
	return s.SendEmail(ctx, fromEmail, toEmail, subject, body, nil, nil)
}

// SendDealerApprovedEmail sends an email to the dealer (booth owner) when their registration is approved. lang: "vi" for Vietnamese, else English.
func (s *MailService) SendDealerApprovedEmail(ctx context.Context, fromEmail, toEmail, boothName, boothNumber, lang string) error {
	var subject, tpl string
	if lang == "vi" {
		subject = "Đơn đăng ký Dealer của bạn đã được duyệt"
		tpl = "dealer_approved_vi.html"
	} else {
		subject = "Your Dealer Registration Has Been Approved"
		tpl = "dealer_approved_en.html"
	}
	data := struct {
		BoothName   string
		BoothNumber string
	}{
		BoothName:   boothName,
		BoothNumber: boothNumber,
	}
	body, err := renderMailTemplate(tpl, data)
	if err != nil {
		return fmt.Errorf("render dealer approved email: %w", err)
	}
	return s.SendEmail(ctx, fromEmail, toEmail, subject, body, nil, nil)
}

// SendTicketApprovedWithQREmail sends an email to the ticket holder when their ticket is approved, with an embedded QR code. lang: "vi" for Vietnamese, else English.
func (s *MailService) SendTicketApprovedWithQREmail(ctx context.Context, fromEmail, toEmail, referenceCode, tierName, lang string) error {
	qrPNG, err := qrcode.Encode(referenceCode, qrcode.Medium, 256)
	if err != nil {
		return fmt.Errorf("generate QR code: %w", err)
	}

	const contentID = "qrcode"
	var subject, tpl string
	tierBlock := htemplate.HTML("")
	if tierName != "" {
		if lang == "vi" {
			tierBlock = htemplate.HTML(fmt.Sprintf(`<p style="margin:0 0 16px 0;font-size:14px;color:#4a4238;"><span style="display:inline-block;background:#fff;padding:8px 14px;border-radius:10px;border:1px solid #dfd5c4;"><strong style="color:#1a1410;">Hạng vé:</strong> %s</span></p>`, htemplate.HTMLEscapeString(tierName)))
		} else {
			tierBlock = htemplate.HTML(fmt.Sprintf(`<p style="margin:0 0 16px 0;font-size:14px;color:#4a4238;"><span style="display:inline-block;background:#fff;padding:8px 14px;border-radius:10px;border:1px solid #dfd5c4;"><strong style="color:#1a1410;">Ticket tier:</strong> %s</span></p>`, htemplate.HTMLEscapeString(tierName)))
		}
	}
	imgSrc := htemplate.URL("cid:" + contentID)
	if lang == "vi" {
		subject = "Vé FUVE chính thức của bạn"
		tpl = "ticket_qr_vi.html"
	} else {
		subject = "Your official FUVE ticket"
		tpl = "ticket_qr_en.html"
	}
	htmlBody, err := renderMailTemplate(tpl, struct {
		ReferenceCode string
		TierBlock     htemplate.HTML
		ImgSrc        htemplate.URL
	}{
		ReferenceCode: referenceCode,
		TierBlock:     tierBlock,
		ImgSrc:        imgSrc,
	})
	if err != nil {
		return fmt.Errorf("render ticket email: %w", err)
	}

	if s.sendgridClient != nil {
		return s.sendEmailSendGridWithInlineImage(ctx, fromEmail, toEmail, subject, htmlBody, contentID, qrPNG)
	}
	if s.sesClient != nil {
		return s.sendEmailSESRawWithInlineImage(ctx, fromEmail, toEmail, subject, htmlBody, contentID, qrPNG)
	}
	return fmt.Errorf("no mail provider configured")
}
