package services

import (
	"github.com/elbuo8/uber-stories/models"
	"github.com/sendgrid/sendgrid-go"
	"log"
	"os"
)

var (
	sg *sendgrid.SGClient
)

func init() {
	sg = sendgrid.NewSendGridClient(os.Getenv("SG_USER"), os.Getenv("SG_PWD"))
}

func ActivationEmail(user *models.User) {
	mail := sendgrid.NewMail()
	mail.AddTo(user.Email)
	mail.AddToName(user.Username)
	mail.SetFrom("yamil@sendgrid.com") // Change later
	mail.SetSubject("Welcome to Uber Stories")
	mail.AddFilter("templates", "enable", "1")
	mail.SetHTML(" ")
	mail.SetText(" ")
	mail.AddFilter("templates", "template_id", "48ac26b1-d586-49f1-8d46-60a0b3d77c6a")
	mail.AddSubstitution("##username##", user.Username)
	mail.AddSubstitution("##verifyURL##", "http://localhost:3000/verify/"+user.ID.Hex()) // Change later
	if err := Send(mail); err != nil {
		log.Println(err)
	}
}

func Send(mail *sendgrid.SGMail) error {
	return sg.Send(mail)
}
