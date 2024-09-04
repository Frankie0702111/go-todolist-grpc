package mail

import (
	"go-todolist-grpc/internal/config"
	"go-todolist-grpc/internal/pkg/log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
)

const (
	CharSet = "UTF-8"
	Region  = "ap-northeast-1"
)

func SendEmail(cnf *config.Config, sender string, recipient []string, bccs []string, subject string, htmlBody string, textBody string) error {

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(Region),
		Credentials: credentials.NewStaticCredentials(cnf.AWSAccessKeyId, cnf.AWSSecretAccessKey, ""),
	})
	if err != nil {
		log.Error.Printf("Error creating AWS session: %v", err)
		return err
	}

	// Create an SES session.
	svc := ses.New(sess)
	bccAddress := []*string{}
	recipientAddress := []*string{}

	for i := range bccs {
		bccAddress = append(bccAddress, &bccs[i])
	}
	for i := range recipient {
		recipientAddress = append(recipientAddress, &recipient[i])
	}

	// Assemble the email.
	html := &ses.Content{}
	if htmlBody != "" {
		html = &ses.Content{
			Charset: aws.String(CharSet),
			Data:    aws.String(htmlBody),
		}
	}

	text := &ses.Content{
		Charset: aws.String(CharSet),
		Data:    aws.String(textBody),
	}

	input := &ses.SendEmailInput{
		Destination: &ses.Destination{
			CcAddresses:  []*string{},
			BccAddresses: bccAddress,
			ToAddresses:  recipientAddress,
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Html: html,
				Text: text,
			},
			Subject: &ses.Content{
				Charset: aws.String(CharSet),
				Data:    aws.String(subject),
			},
		},
		Source: aws.String(sender),
	}

	// Attempt to send the email.
	result, err := svc.SendEmail(input)

	// Display error messages if they occur.
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case ses.ErrCodeMessageRejected:
				log.Error.Print(ses.ErrCodeMessageRejected, aerr.Error())
			case ses.ErrCodeMailFromDomainNotVerifiedException:
				log.Error.Print(ses.ErrCodeMailFromDomainNotVerifiedException, aerr.Error())
			case ses.ErrCodeConfigurationSetDoesNotExistException:
				log.Error.Print(ses.ErrCodeConfigurationSetDoesNotExistException, aerr.Error())
			default:
				log.Error.Print(aerr.Error())
			}
		} else {
			log.Error.Print(err.Error())
		}

		return err
	}
	log.Debug.Printf("Email Sent to address: %s, result: %v", recipient, result)

	return nil
}
