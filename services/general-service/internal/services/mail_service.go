package services

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"general-service/internal/repositories"
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
		Source: aws.String(fromEmail),
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
	var subject, body string
	if lang == "vi" {
		subject = "Xác thực email của bạn cho FUVE"
		body = fmt.Sprintf(`
		<html>
			<body style="font-family: Arial, sans-serif; margin: 0; padding: 0; background-color: #f0e68c;">
				<div style="max-width: 640px; margin: 0 auto; padding: 24px; box-sizing: border-box;">
					<div style="background-color: #ffffff; border: 12px solid #e6c200; border-radius: 4px; padding: 30px; margin: 0;">
						<p style="color: #333; font-size: 16px; margin: 0 0 16px 0;"><strong>Kính gửi Người tham gia,</strong></p>
						<p style="color: #333; font-size: 16px; margin: 0 0 12px 0;">Cảm ơn bạn đã quan tâm đến FUVE. Để tiếp tục và mang đến trải nghiệm tốt nhất, vui lòng xác thực email bằng mã bên dưới.</p>
						<p style="color: #333; font-size: 14px; margin: 16px 0 8px 0;"><strong>Mã xác thực của bạn:</strong></p>
						<div style="background-color: #f9f9f9; padding: 20px; text-align: center; border-radius: 6px; margin: 12px 0 20px 0;">
							<span style="color: #333; font-size: 28px; letter-spacing: 6px; font-weight: bold;">%s</span>
						</div>
						<p style="color: #666; font-size: 14px; margin: 0 0 8px 0;">Mã này hết hạn sau 10 phút.</p>
						<p style="color: #333; font-size: 14px; margin: 20px 0 8px 0;">Xác thực giúp chúng tôi:</p>
						<ul style="color: #333; font-size: 14px; margin: 0 0 20px 0; padding-left: 20px;">
							<li>Gửi vé và thông tin cập nhật đúng địa chỉ của bạn</li>
							<li>Thông báo cho bạn về sự kiện</li>
							<li>Xác nhận đăng ký của bạn trong hệ thống</li>
						</ul>
						<p style="color: #999; font-size: 12px; margin: 0;">Nếu bạn không yêu cầu, hãy bỏ qua email này.</p>
						<p style="color: #333; font-size: 14px; margin: 24px 0 0 0;">Chúng tôi rất vui được đón bạn và hẹn gặp tại FUVE.</p>
						<p style="color: #333; font-size: 14px; margin: 24px 0 0 0;">Trân trọng,<br><strong>FUVE</strong></p>
						<hr style="border: none; border-top: 1px solid #ddd; margin: 28px 0 16px 0;">
						<p style="color: #666; font-size: 12px; margin: 0;">Liên hệ: fuve.vn &middot; Facebook: FUVE - Furry Vietnam Eternity<br>Hẹn gặp bạn tại FUVE!</p>
					</div>
				</div>
			</body>
		</html>
	`, otp)
	} else {
		subject = "Verify your email for FUVE"
		body = fmt.Sprintf(`
		<html>
			<body style="font-family: Arial, sans-serif; margin: 0; padding: 0; background-color: #f0e68c;">
				<div style="max-width: 640px; margin: 0 auto; padding: 24px; box-sizing: border-box;">
					<div style="background-color: #ffffff; border: 12px solid #e6c200; border-radius: 4px; padding: 30px; margin: 0;">
						<p style="color: #333; font-size: 16px; margin: 0 0 16px 0;"><strong>Dear Participant,</strong></p>
						<p style="color: #333; font-size: 16px; margin: 0 0 12px 0;">Thanks for your interest in FUVE. To continue and so we can give you the best experience, please verify your email using the code below.</p>
						<p style="color: #333; font-size: 14px; margin: 16px 0 8px 0;"><strong>Your verification code:</strong></p>
						<div style="background-color: #f9f9f9; padding: 20px; text-align: center; border-radius: 6px; margin: 12px 0 20px 0;">
							<span style="color: #333; font-size: 28px; letter-spacing: 6px; font-weight: bold;">%s</span>
						</div>
						<p style="color: #666; font-size: 14px; margin: 0 0 8px 0;">This code expires in 10 minutes.</p>
						<p style="color: #333; font-size: 14px; margin: 20px 0 8px 0;">Verifying helps us:</p>
						<ul style="color: #333; font-size: 14px; margin: 0 0 20px 0; padding-left: 20px;">
							<li>Send your ticket and updates to the right address</li>
							<li>Keep you informed about the event</li>
							<li>Confirm your registration in our system</li>
						</ul>
						<p style="color: #999; font-size: 12px; margin: 0;">If you didn't request this, you can ignore this email.</p>
						<p style="color: #333; font-size: 14px; margin: 24px 0 0 0;">We're excited to have you and look forward to seeing you at FUVE.</p>
						<p style="color: #333; font-size: 14px; margin: 24px 0 0 0;">Sincerely,<br><strong>FUVE</strong></p>
						<hr style="border: none; border-top: 1px solid #ddd; margin: 28px 0 16px 0;">
						<p style="color: #666; font-size: 12px; margin: 0;">Contact us: fuve.vn &middot; Facebook: FUVE - Furry Vietnam Eternity<br>See you soon at FUVE!</p>
					</div>
				</div>
			</body>
		</html>
	`, otp)
	}
	return s.SendEmail(ctx, fromEmail, toEmail, subject, body, nil, nil)
}

// SendDealerApprovedEmail sends an email to the dealer (booth owner) when their registration is approved. lang: "vi" for Vietnamese, else English.
func (s *MailService) SendDealerApprovedEmail(ctx context.Context, fromEmail, toEmail, boothName, boothNumber, lang string) error {
	var subject, body string
	wrap := `style="font-family: Arial, sans-serif; margin: 0; padding: 0; background-color: #f0e68c;"`
	inner := `style="background-color: #ffffff; border: 12px solid #e6c200; border-radius: 4px; padding: 30px; margin: 0;"`
	if lang == "vi" {
		subject = "Đơn đăng ký Dealer của bạn đã được duyệt"
		body = fmt.Sprintf(`
		<html><body %s>
			<div style="max-width: 640px; margin: 0 auto; padding: 24px; box-sizing: border-box;">
				<div %s>
					<p style="color: #333; font-size: 16px; margin: 0 0 16px 0;"><strong>Kính gửi Người tham gia,</strong></p>
					<p style="color: #333; font-size: 16px; margin: 0 0 12px 0;">Đơn đăng ký gian hàng dealer của bạn đã được xác minh và phê duyệt.</p>
					<div style="background-color: #f9f9f9; padding: 20px; border-radius: 6px; margin: 20px 0;">
						<p style="color: #333; font-size: 14px; margin: 0 0 8px 0;"><strong>Tên gian hàng:</strong> %s</p>
						<p style="color: #333; font-size: 14px; margin: 0 0 8px 0;"><strong>Mã gian hàng / Số booth:</strong> <code style="background: #eee; padding: 4px 8px; border-radius: 4px;">%s</code></p>
						<p style="color: #666; font-size: 13px; margin: 12px 0 0 0;">Chia sẻ mã gian hàng này với nhân viên để họ tham gia gian hàng của bạn.</p>
					</div>
					<p style="color: #999; font-size: 12px; margin-top: 20px;">Nếu có thắc mắc, vui lòng liên hệ ban tổ chức.</p>
					<p style="color: #333; font-size: 14px; margin: 24px 0 0 0;">Trân trọng,<br><strong>FUVE</strong></p>
					<hr style="border: none; border-top: 1px solid #ddd; margin: 28px 0 16px 0;">
					<p style="color: #666; font-size: 12px; margin: 0;">Liên hệ: fuve.vn &middot; Facebook: FUVE - Furry Vietnam Eternity<br>Hẹn gặp bạn tại FUVE!</p>
				</div>
			</div>
		</body></html>
		`, wrap, inner, boothName, boothNumber)
	} else {
		subject = "Your Dealer Registration Has Been Approved"
		body = fmt.Sprintf(`
		<html><body %s>
			<div style="max-width: 640px; margin: 0 auto; padding: 24px; box-sizing: border-box;">
				<div %s>
					<p style="color: #333; font-size: 16px; margin: 0 0 16px 0;"><strong>Dear Participant,</strong></p>
					<p style="color: #333; font-size: 16px; margin: 0 0 12px 0;">Your dealer booth registration has been verified and approved.</p>
					<div style="background-color: #f9f9f9; padding: 20px; border-radius: 6px; margin: 20px 0;">
						<p style="color: #333; font-size: 14px; margin: 0 0 8px 0;"><strong>Booth name:</strong> %s</p>
						<p style="color: #333; font-size: 14px; margin: 0 0 8px 0;"><strong>Dealer den / Booth number:</strong> <code style="background: #eee; padding: 4px 8px; border-radius: 4px;">%s</code></p>
						<p style="color: #666; font-size: 13px; margin: 12px 0 0 0;">Share this booth code with staff so they can join your dealer booth.</p>
					</div>
					<p style="color: #999; font-size: 12px; margin-top: 20px;">If you have any questions, please contact the event organizers.</p>
					<p style="color: #333; font-size: 14px; margin: 24px 0 0 0;">Sincerely,<br><strong>FUVE</strong></p>
					<hr style="border: none; border-top: 1px solid #ddd; margin: 28px 0 16px 0;">
					<p style="color: #666; font-size: 12px; margin: 0;">Contact us: fuve.vn &middot; Facebook: FUVE - Furry Vietnam Eternity<br>See you soon at FUVE!</p>
				</div>
			</div>
		</body></html>
		`, wrap, inner, boothName, boothNumber)
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
	var subject, htmlBody string
	tierLineEn := ""
	tierLineVi := ""
	if tierName != "" {
		tierLineEn = fmt.Sprintf(`<p style="color: #333; font-size: 14px; margin: 0 0 8px 0;"><strong>Ticket tier:</strong> %s</p>`, tierName)
		tierLineVi = fmt.Sprintf(`<p style="color: #333; font-size: 14px; margin: 0 0 8px 0;"><strong>Hạng vé:</strong> %s</p>`, tierName)
	}
	if lang == "vi" {
		subject = "Vé FUVE chính thức của bạn"
		htmlBody = fmt.Sprintf(`
		<html>
			<body style="font-family: Arial, sans-serif; margin: 0; padding: 0; background-color: #f0e68c;">
				<div style="max-width: 640px; margin: 0 auto; padding: 24px; box-sizing: border-box;">
					<div style="background-color: #ffffff; border: 12px solid #e6c200; border-radius: 4px; padding: 30px; margin: 0;">
						<p style="color: #333; font-size: 16px; margin: 0 0 16px 0;"><strong>Kính gửi Người tham gia,</strong></p>
						<p style="color: #333; font-size: 16px; margin: 0 0 12px 0;">Đây là <strong>vé FUVE chính thức</strong> của bạn. Cảm ơn bạn đã là một phần của cộng đồng FUVE. Vé của bạn được đính kèm bên dưới và là chìa khóa để vào sự kiện.</p>
						<p style="color: #333; font-size: 14px; margin: 16px 0 8px 0;"><strong>Lưu ý:</strong></p>
						<ul style="color: #333; font-size: 14px; margin: 0 0 20px 0; padding-left: 20px;">
							<li>Mang theo vé này (bản điện tử hoặc in) trong ngày diễn ra sự kiện</li>
							<li>Mỗi vé chỉ có hiệu lực cho <strong>một người tham gia</strong></li>
							<li>Giữ email này để làm thủ tục check-in và xác minh</li>
						</ul>
						<div style="text-align: center; margin: 24px 0;">
							<p style="color: #333; font-size: 12px; margin: 0 0 8px 0;">Mã tham chiếu: <code style="background: #eee; padding: 2px 6px; border-radius: 4px;">%s</code></p>
							%s
							<p style="margin: 12px 0 0 0;"><img src="cid:%s" alt="Mã QR vé" width="200" height="200" style="display: block; margin: 0 auto;" /></p>
						</div>
						<p style="color: #666; font-size: 14px; margin: 24px 0 0 0;">Nếu có thắc mắc về vé, lịch trình hoặc tham gia sự kiện, hãy liên hệ với chúng tôi bất cứ lúc nào.</p>
						<p style="color: #333; font-size: 14px; margin: 16px 0 0 0;">Hẹn gặp bạn tại sự kiện. Cùng tạo nên một FUVE đáng nhớ.</p>
						<p style="color: #333; font-size: 14px; margin: 24px 0 0 0;">Trân trọng,<br><strong>FUVE</strong></p>
						<hr style="border: none; border-top: 1px solid #ddd; margin: 28px 0 16px 0;">
						<p style="color: #666; font-size: 12px; margin: 0;">Liên hệ: fuve.vn &middot; Facebook: FUVE - Furry Vietnam Eternity<br>Hẹn gặp bạn tại FUVE!</p>
					</div>
				</div>
			</body>
		</html>
		`, referenceCode, tierLineVi, contentID)
	} else {
		subject = "Your official FUVE ticket"
		htmlBody = fmt.Sprintf(`
		<html>
			<body style="font-family: Arial, sans-serif; margin: 0; padding: 0; background-color: #f0e68c;">
				<div style="max-width: 640px; margin: 0 auto; padding: 24px; box-sizing: border-box;">
					<div style="background-color: #ffffff; border: 12px solid #e6c200; border-radius: 4px; padding: 30px; margin: 0;">
						<p style="color: #333; font-size: 16px; margin: 0 0 16px 0;"><strong>Dear Participant,</strong></p>
						<p style="color: #333; font-size: 16px; margin: 0 0 12px 0;">This is your <strong>official FUVE ticket</strong>. Thank you for being part of the FUVE community. Your ticket is included below and will be your key to entering the event.</p>
						<p style="color: #333; font-size: 14px; margin: 16px 0 8px 0;"><strong>Please note:</strong></p>
						<ul style="color: #333; font-size: 14px; margin: 0 0 20px 0; padding-left: 20px;">
							<li>Bring this ticket (digital or printed) on event day</li>
							<li>Each ticket is valid for <strong>one participant only</strong></li>
							<li>Keep this email for check-in and verification</li>
						</ul>
						<div style="text-align: center; margin: 24px 0;">
							<p style="color: #333; font-size: 12px; margin: 0 0 8px 0;">Reference code: <code style="background: #eee; padding: 2px 6px; border-radius: 4px;">%s</code></p>
							%s
							<p style="margin: 12px 0 0 0;"><img src="cid:%s" alt="QR Ticket Code" width="200" height="200" style="display: block; margin: 0 auto;" /></p>
						</div>
						<p style="color: #666; font-size: 14px; margin: 24px 0 0 0;">If you have any questions regarding your ticket, schedule, or participation, feel free to reach out to us anytime.</p>
						<p style="color: #333; font-size: 14px; margin: 16px 0 0 0;">See you there. Let's create an unforgettable FUVE together.</p>
						<p style="color: #333; font-size: 14px; margin: 24px 0 0 0;">Sincerely,<br><strong>FUVE</strong></p>
						<hr style="border: none; border-top: 1px solid #ddd; margin: 28px 0 16px 0;">
						<p style="color: #666; font-size: 12px; margin: 0;">Contact us: fuve.vn &middot; Facebook: FUVE - Furry Vietnam Eternity<br>See you soon at FUVE!</p>
					</div>
				</div>
			</body>
		</html>
		`, referenceCode, tierLineEn, contentID)
	}

	if s.sendgridClient != nil {
		return s.sendEmailSendGridWithInlineImage(ctx, fromEmail, toEmail, subject, htmlBody, contentID, qrPNG)
	}
	if s.sesClient != nil {
		return s.sendEmailSESRawWithInlineImage(ctx, fromEmail, toEmail, subject, htmlBody, contentID, qrPNG)
	}
	return fmt.Errorf("no mail provider configured")
}
