package email

import (
	"github.com/jordan-wright/email"
	"github.com/duncan/config"
	"net/smtp"
	"fmt"
	"log"
	"strings"
)

func Alert(subject, body string) {
	cfg := config.EmailConfig()
	recipients := strings.Split(cfg.Alert.Recipients, ",")
	Email(recipients, subject, body)
}

func Email(recipients []string, subject, body string) {
	cfg := config.EmailConfig()
	e := email.NewEmail()
	e.From = fmt.Sprintf("%s <%s>", "Nebulo Backend Email Bot", cfg.Alert.Address)
	e.To = recipients
	e.Subject = subject
	e.Text = []byte(body)
	err := e.Send(fmt.Sprintf("%s:%s", cfg.Alert.SMTPServer, cfg.Alert.SMTPPort), smtp.PlainAuth("", cfg.Alert.Address, cfg.Alert.Password, cfg.Alert.SMTPServer))
	if err != nil {
		log.Fatal(err)
	}
}