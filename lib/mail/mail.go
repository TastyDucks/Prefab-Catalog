package mail

import (
	"Prefab-Catalog/lib/config"
	"Prefab-Catalog/lib/lumberjack"
	"fmt"
	"os/exec"
	"time"

	"net/smtp"

	"github.com/matcornic/hermes/v2"
)

var log *lumberjack.Lumberjack = lumberjack.New("Mail")

/*
Send sends an email with the contents of the order specified by orderID.
*/
func Send(targets []string, orderID string) {
	// Generate email.
	url := config.Load().URL
	from := "mail@" + url
	subject := "Prefab Catalog - New Order"
	h := hermes.Hermes{
		Theme:         nil,
		TextDirection: "",
		Product: hermes.Product{
			Name:        "Prefab Catalog",
			Link:        url,
			Logo:        "http://" + url + "/media/logo_color.png",
			Copyright:   "© 2021 PKRE.CO",
			TroubleText: "",
		},
		DisableCSSInlining: false,
	}
	email := hermes.Email{
		Body: hermes.Body{
			Name:         "",
			Intros:       []string{"A new prefab order has been created."},
			Dictionary:   []hermes.Entry{},
			Table:        hermes.Table{},
			Actions:      []hermes.Action{{Button: hermes.Button{Text: "View order", Link: url + "/order/" + orderID}}},
			Outros:       []string{},
			Greeting:     "",
			Signature:    fmt.Sprintf("Email generated at %s", time.Now().Format("2006-01-02 15:04:05")),
			Title:        "Prefab Catalog - New Order",
			FreeMarkdown: "",
		},
	}
	emailBody, err := h.GenerateHTML(email)
	if err != nil {
		log.Error(err, "failed to generate email body")
	}
	for index, target := range targets {
		log.Debugf("Attempting to send email %d of %d (FROM: %s, TO: %s)", index+1, len(targets), from, target)
		emailFull := []byte(fmt.Sprintf("From: %v\r\nTo: %v\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=\"utf-8\"\r\n\r\n%s", from, target, subject, emailBody))

		c, err := smtp.Dial("prefab.pkre.co:7734") // TODO: Load from config and run as goroutine so order submit page isn't blocked by connection timeout.
		if err != nil {
			log.Error(err, "failed to connect to smtp server")
			return
		}
		defer c.Close()
		if err = c.Mail(from); err != nil {
			log.Error(err, "failed to set smpt from")
		}
		if err = c.Rcpt(target); err != nil {
			log.Error(err, "failed to set smpt rcpt")
		}
		w, err := c.Data()
		if err != nil {
			log.Error(err, "failed to get data object from smtp server")
			return
		}
		if _, err = w.Write(emailFull); err != nil {
			log.Error(err, "failed to stream email body")
			return
		}
		if err = w.Close(); err != nil {
			log.Error(err, "failed to close smtp writer")
		}
		if err = c.Quit(); err != nil {
			log.Error(err, "failed to quit smtp connection")
		}
	}
	return
}

/*
LogGet returns the contents of postfix's mail log.mai
*/
func LogGet() string {
	var stdout []byte
	var err error
	if stdout, err = exec.Command("grep", "", "/var/log/mail.log").Output(); err != nil {
		log.Error(err, "failed to get mail log")
		return err.Error()
	}
	return string(stdout)
}
