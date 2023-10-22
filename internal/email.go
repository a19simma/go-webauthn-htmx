package internal

import (
	"context"
	"fmt"

	brevo "github.com/getbrevo/brevo-go/lib"
	"github.com/rs/zerolog/log"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/spf13/viper"
)

func SendEmail(recipientEmail string, recipientName string, link string) error {
	apiKey := viper.Get("SENDGRID_API_KEY").(string)

	from := mail.NewEmail("emailsender", "emailsender@simonmalm.com")
	subject := "Registration"
	to := mail.NewEmail(recipientName, recipientEmail)
	plainTextContent := fmt.Sprintf("To complete the registration go to link: %v", link)
	htmlContent := ""
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
	client := sendgrid.NewSendClient(apiKey)
	response, err := client.Send(message)
	if err != nil {
		return brevoFallback(recipientEmail, recipientName, link, err)
	}

	log.Debug().Msgf("Succeeded in sending email: %v", response)
	return nil
}

func brevoFallback(recipientEmail string, recipientName string, link string, incomingErr error) error {
	apiKey := viper.Get("BREVO_API_KEY").(string)
	var ctx context.Context
	cfg := brevo.NewConfiguration()

	cfg.AddDefaultHeader("api-key", apiKey)

	br := brevo.NewAPIClient(cfg)

	message := brevo.SendSmtpEmail{
		Sender: &brevo.SendSmtpEmailSender{
			Name:  "emailsender",
			Email: "emailsender@simonmalm.com",
		},
		To: []brevo.SendSmtpEmailTo{{
			Email: recipientEmail,
			Name:  recipientName,
		}},
		HtmlContent: fmt.Sprintf("To complete the registration go to link: %v", link),
		TextContent: "",
		Subject:     "Registration",
		ReplyTo: &brevo.SendSmtpEmailReplyTo{
			Email: "no-reply@simonmalm.com",
			Name:  "no-reply",
		},
		Attachment: []brevo.SendSmtpEmailAttachment{},
	}
	result, resp, err := br.TransactionalEmailsApi.SendTransacEmail(ctx, message)
	if err != nil {
		log.Error().Msgf("Error when calling AccountApi->get_account: %v : %v",
			incomingErr.Error(), err.Error())
		return err
	}

	log.Debug().Msgf("Successfully send email with result: %v and response: %v", result, resp)
	return nil
}
