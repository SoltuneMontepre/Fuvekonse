package services

import (
	"context"
	"fmt"
	"general-service/internal/repositories"

	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
)

type MailService struct {
	repos     *repositories.Repositories
	sesClient *ses.Client
}

func NewMailService(repos *repositories.Repositories) *MailService {
	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("unable to load AWS SDK config: %v", err)
	}

	client := ses.NewFromConfig(cfg)

	return &MailService{
		repos:     repos,
		sesClient: client,
	}
}

func (s *MailService) SendEmail(ctx context.Context, fromEmail, toEmail, subject, body string, cc, bcc []string) error {
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
