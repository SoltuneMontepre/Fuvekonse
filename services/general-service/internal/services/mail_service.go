package services

import (
	"context"
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
)

const (
	mailProviderSES      = "ses"
	mailProviderSendGrid = "sendgrid"
)

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

// SendOtpEmail sends an OTP verification email to the specified email address
func (s *MailService) SendOtpEmail(ctx context.Context, fromEmail, toEmail, otp string) error {
	subject := "Email Verification OTP"
	body := fmt.Sprintf(`
		<html>
			<body style="font-family: Arial, sans-serif; padding: 20px;">
				<div style="max-width: 600px; margin: 0 auto; background-color: #f5f5f5; padding: 30px; border-radius: 10px;">
					<h2 style="color: #333; text-align: center;">Email Verification</h2>
					<p style="color: #666; font-size: 16px;">Your OTP for email verification is:</p>
					<div style="background-color: #fff; padding: 20px; text-align: center; border-radius: 5px; margin: 20px 0;">
						<h1 style="color: #007bff; letter-spacing: 5px; margin: 0;">%s</h1>
					</div>
					<p style="color: #666; font-size: 14px;">This code will expire in 10 minutes.</p>
					<p style="color: #999; font-size: 12px; margin-top: 30px;">If you didn't request this code, please ignore this email.</p>
				</div>
			</body>
		</html>
	`, otp)

	return s.SendEmail(ctx, fromEmail, toEmail, subject, body, nil, nil)
}

// SendDealerApprovedEmail sends an email to the dealer (booth owner) when their registration is approved, including dealer den (booth) information.
func (s *MailService) SendDealerApprovedEmail(ctx context.Context, fromEmail, toEmail, boothName, boothNumber string) error {
	subject := "Your Dealer Registration Has Been Approved"
	body := fmt.Sprintf(`
		<html>
			<body style="font-family: Arial, sans-serif; padding: 20px;">
				<div style="max-width: 600px; margin: 0 auto; background-color: #f5f5f5; padding: 30px; border-radius: 10px;">
					<h2 style="color: #333; text-align: center;">Dealer Registration Approved</h2>
					<p style="color: #666; font-size: 16px;">Your dealer booth registration has been verified and approved.</p>
					<div style="background-color: #fff; padding: 20px; border-radius: 5px; margin: 20px 0;">
						<p style="color: #333; font-size: 14px; margin: 0 0 8px 0;"><strong>Booth name:</strong> %s</p>
						<p style="color: #333; font-size: 14px; margin: 0 0 8px 0;"><strong>Dealer den / Booth number:</strong> <code style="background: #eee; padding: 4px 8px; border-radius: 4px;">%s</code></p>
						<p style="color: #666; font-size: 13px; margin: 12px 0 0 0;">Share this booth code with staff so they can join your dealer booth.</p>
					</div>
					<p style="color: #999; font-size: 12px; margin-top: 30px;">If you have any questions, please contact the event organizers.</p>
				</div>
			</body>
		</html>
	`, boothName, boothNumber)

	return s.SendEmail(ctx, fromEmail, toEmail, subject, body, nil, nil)
}
