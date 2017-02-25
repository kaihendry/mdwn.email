package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/tidwall/gjson"

	"github.com/apex/go-apex"
	apexSNS "github.com/apex/go-apex/sns"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/jhillyerd/enmime"
)

func main() {
	apexSNS.HandleFunc(func(event *apexSNS.Event, ctx *apex.Context) error {
		// fmt.Fprintln(os.Stderr, event.Records[0].SNS.Message)

		messageId := gjson.Get(event.Records[0].SNS.Message, "mail.messageId").String()
		from := gjson.Get(event.Records[0].SNS.Message, "mail.source").String()
		subject := gjson.Get(event.Records[0].SNS.Message, "mail.commonHeaders.subject").String()
		fmt.Fprintln(os.Stderr, messageId)

		sess := session.Must(session.NewSession())
		s3svc := s3.New(sess)
		sessvc := ses.New(sess)

		// https://docs.aws.amazon.com/sdk-for-go/api/service/s3/#example_S3_GetObject
		getparams := &s3.GetObjectInput{
			Bucket: aws.String("mdwn-inbox"), // Required
			Key:    aws.String(messageId),    // Required
		}

		getresp, err := s3svc.GetObject(getparams)

		if err != nil {
			return err
		}

		env, err := enmime.ReadEnvelope(getresp.Body)
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "From: %v\n", env.GetHeader("From"))

		// Pretty-print the response data.
		fmt.Fprintln(os.Stderr, getresp)
		// b, _ := ioutil.ReadAll(getresp.Body)

		// https://docs.aws.amazon.com/sdk-for-go/api/service/s3/#PutObjectInput
		// https://docs.aws.amazon.com/sdk-for-go/api/service/s3/#example_S3_PutObject
		putparams := &s3.PutObjectInput{
			Bucket:       aws.String("mdwn-web"),            // Required
			Body:         bytes.NewReader([]byte(env.Text)), // Required
			Key:          aws.String(messageId + ".txt"),    // Required
			ContentType:  aws.String("text/plain; charset=UTF-8"),
			ACL:          aws.String("public-read"),
			StorageClass: aws.String("REDUCED_REDUNDANCY"),
		}

		putresp, err := s3svc.PutObject(putparams)

		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "Managed to upload", putresp)

		// Email

		mailparams := &ses.SendEmailInput{
			Destination: &ses.Destination{
				ToAddresses: []*string{
					aws.String(from),
				},
			},
			Message: &ses.Message{
				Body: &ses.Body{
					Text: &ses.Content{
						Data: aws.String("https://mdwn.email/" + messageId + ".txt"),
					},
				},
				Subject: &ses.Content{
					Data: aws.String("RE: " + subject),
				},
			},
			Source: aws.String("txt@mdwn.email"),
		}

		_, err = sessvc.SendEmail(mailparams)

		if err != nil {
			return err
		}

		return nil
	})
}
