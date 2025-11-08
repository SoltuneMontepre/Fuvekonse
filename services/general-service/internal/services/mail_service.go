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
	repos *repositories.Repositories
}

func NewMailService(repos *repositories.Repositories) *MailService {
	return &MailService{
		repos: repos,
	}
}

func (s *MailService) SendEmail(fromEmail, toEmail, subject, body string, cc, bcc []string) error {
	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("unable to load SDK config: %w", err)
	}

	client := ses.NewFromConfig(cfg)

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

	resp, err := client.SendEmail(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	log.Printf("Email sent successfully! Message ID: %s\n", *resp.MessageId)
	return nil
}
