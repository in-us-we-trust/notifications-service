package mailtrap

import (
	"errors"
	"fmt"

	gomail "gopkg.in/mail.v2"
)

type Client struct {
	Config Config
}

type Config struct {
	FromPsw   string
	FromEmail string
}

func (c *Config) Validate() error {
	if c.FromPsw == "" {
		return errors.New("from password is required")
	}
	if c.FromEmail == "" {
		return errors.New("from email is required")
	}
	return nil
}

func NewMailTrapClient(conf Config) (*Client, error) {
	if err := conf.Validate(); err != nil {
		return nil, err
	}

	client := &Client{
		Config: conf,
	}

	return client, nil
}

func (o *Client) Send(email, subject, body string) error {
	message := gomail.NewMessage()
	message.SetHeader("From", fmt.Sprintf("%s <%s>", "Vigil-IA", o.Config.FromEmail))
	message.SetHeader("To", email)
	message.SetHeader("Subject", subject)
	message.AddAlternative("text/html", body)

	dialer := gomail.NewDialer("smtp.gmail.com", 587, o.Config.FromEmail, o.Config.FromPsw)

	if err := dialer.DialAndSend(message); err != nil {
		return err
	}

	return nil
}

func (o *Client) SendBulk(emails []string, subject, body string) error {
	message := gomail.NewMessage()
	message.SetHeader("From", fmt.Sprintf("%s <%s>", "Vigil-IA", o.Config.FromEmail))
	message.SetHeader("Bcc", emails...)
	message.SetHeader("Subject", subject)
	message.AddAlternative("text/html", body)

	dialer := gomail.NewDialer("smtp.gmail.com", 587, o.Config.FromEmail, o.Config.FromPsw)

	if err := dialer.DialAndSend(message); err != nil {
		return err
	}

	return nil
}
