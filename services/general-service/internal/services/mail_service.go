package services

import (
	"context"
	"fmt"
	"general-service/internal/repositories"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
)

type MailService struct {
	repos     *repositories.Repositories
	sesClient *ses.Client
}

func NewMailService(repos *repositories.Repositories) *MailService {
	ctx := context.Background()

	region := os.Getenv("AWS_REGION")

	useLocalStack := os.Getenv("USE_LOCALSTACK") == "true"
	var cfg aws.Config
	var err error

	if useLocalStack {
		accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
		secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
		if accessKey == "" || secretKey == "" {
			log.Fatalf("USE_LOCALSTACK=true but AWS_ACCESS_KEY_ID or AWS_SECRET_ACCESS_KEY not set")
		}

		localEndpoint := os.Getenv("LOCALSTACK_ENDPOINT")

		cfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(region),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
			config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{
					URL:           localEndpoint,
					SigningRegion: region,
				}, nil
			})),
		)
	} else {
		cfg, err = config.LoadDefaultConfig(ctx, config.WithRegion(region))
	}

	if err != nil {
		log.Fatalf("failed to load AWS SDK config: %v", err)
	}

	client := ses.NewFromConfig(cfg)
	return &MailService{repos: repos, sesClient: client}
}

func (s *MailService) SendEmail(ctx context.Context, fromEmail, toEmail, subject, body string, cc, bcc []string) error {
	// log the email content ...
	if os.Getenv("USE_LOCALSTACK") == "true" {
		log.Printf("[EMAIL CONTENT] From: %s To: %s Subject: %s\nBody:\n%s\n", fromEmail, toEmail, subject, body)
	}

	if s.sesClient == nil {
		return fmt.Errorf("SES client not configured")
	}

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
		log.Printf("failed to send email via SES: %v (from=%s to=%s)", err, fromEmail, toEmail)
		return fmt.Errorf("failed to send email: %w", err)
	}

	log.Printf("Email sent successfully! Message ID: %s\n", *resp.MessageId)
	return nil
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
