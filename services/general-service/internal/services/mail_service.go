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
