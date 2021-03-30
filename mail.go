// using SendGrid's Go Library
// https://github.com/sendgrid/sendgrid-go
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

func SendMail(sh Share) error {
	if len(sh.Emails) <= 0 {
		return nil
	}
	m := mail.NewV3Mail()
	m.SetFrom(mail.NewEmail(os.Getenv("SENDGRID_SENDER_NAME"), os.Getenv("SENDGRID_SENDER_MAIL")))
	m.SetTemplateID(os.Getenv("SENDGRID_SHARE_TEMPLATE"))
	p := mail.NewPersonalization()

	var receivers []*mail.Email
	for _, address := range sh.Emails {
		receivers = append(receivers, mail.NewEmail(strings.Split(address, "@")[0], address))
	}
	fmt.Println(sh.Emails)
	p.AddTos(receivers...)
	p.SetDynamicTemplateData("id", sh.ID.String())
	p.SetDynamicTemplateData("download_id", sh.DownloadLimit)
	p.SetDynamicTemplateData("files", sh.Attachments)
	m.AddPersonalizations(p)

	request := sendgrid.GetRequest(os.Getenv("SENDGRID_API_KEY"), "/v3/mail/send", "https://api.sendgrid.com")
	request.Method = "POST"
	request.Body = mail.GetRequestBody(m)
	_, err := sendgrid.API(request)
	return err
}