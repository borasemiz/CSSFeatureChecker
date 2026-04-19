package email

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	sestypes "github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/caniuse-scraper/scraper"
)

type SESv2SendEmailAPI interface {
	SendEmail(ctx context.Context, input *sesv2.SendEmailInput, opts ...func(*sesv2.Options)) (*sesv2.SendEmailOutput, error)
}

func Send(ctx context.Context, client SESv2SendEmailAPI, from, to string, features []scraper.Result) error {
	var sb strings.Builder
	sb.WriteString("The following CSS features have newly crossed 90% browser coverage:\n\n")
	for _, f := range features {
		fmt.Fprintf(&sb, "- %s: %.2f%%\n  %s\n\n", f.Title, f.Coverage, f.URL)
	}

	_, err := client.SendEmail(ctx, &sesv2.SendEmailInput{
		FromEmailAddress: aws.String(from),
		Destination: &sestypes.Destination{
			ToAddresses: []string{to},
		},
		Content: &sestypes.EmailContent{
			Simple: &sestypes.Message{
				Subject: &sestypes.Content{
					Data: aws.String("CSS Features Newly Above 90% Coverage"),
				},
				Body: &sestypes.Body{
					Text: &sestypes.Content{
						Data: aws.String(sb.String()),
					},
				},
			},
		},
	})
	return err
}
