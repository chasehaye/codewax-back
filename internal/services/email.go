package services

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
)

var (
	sesClient     *ses.Client
	senderAddress string
)

func init() {
	senderAddress = os.Getenv("SENDER_ADDRESS")

	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion(os.Getenv("AWS_DEFAULT_REGION")),
	)
	if err != nil {
		panic(fmt.Errorf("aws config init error: %w", err))
	}

	sesClient = ses.NewFromConfig(cfg)
}

func SendMail(to, subject, body string) error {
	input := &ses.SendEmailInput{
		Source: &senderAddress,
		Destination: &types.Destination{
			ToAddresses: []string{to},
		},
		Message: &types.Message{
			Subject: &types.Content{
				Data: &subject,
			},
			Body: &types.Body{
				Html: &types.Content{
					Data: &body,
				},
			},
		},
	}

	_, err := sesClient.SendEmail(context.TODO(), input)

	if err != nil {
		fmt.Printf("SES ERROR: %v\n", err)
		return fmt.Errorf("ses send failed: %w", err)
	}
	
	return nil
}