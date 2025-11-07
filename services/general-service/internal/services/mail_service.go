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

func (s *MailService) SendEmail(fromEmail string) {
    ctx := context.Background()

    cfg, err := config.LoadDefaultConfig(ctx)
    if err != nil {
        log.Fatalf("unable to load SDK config: %v", err)
    }

    client := ses.NewFromConfig(cfg)

    input := &ses.SendEmailInput{
        Source: aws.String(fromEmail),
        Destination: &types.Destination{
            ToAddresses: []string{"recipient@example.com"},
        },
        Message: &types.Message{
            Subject: &types.Content{
                Data:    aws.String("Test email from AWS SES via Go SDK"),
                Charset: aws.String("UTF-8"),
            },
            Body: &types.Body{
                Text: &types.Content{
                    Data:    aws.String("This is the plain-text portion of the email."),
                    Charset: aws.String("UTF-8"),
                },
                Html: &types.Content{
                    Data:    aws.String("<html><body><h1>Hello!</h1><p>This is a HTML email.</p></body></html>"),
                    Charset: aws.String("UTF-8"),
                },
            },
        },
    }

    resp, err := client.SendEmail(ctx, input)
    if err != nil {
        log.Fatalf("failed to send email: %v", err)
    }

    fmt.Printf("Email sent! Message ID: %s\n", *resp.MessageId)
}