package mailer

import (
	"bytes"
	"embed"
	"notification-service/lib/mailtrap"
	"text/template"
)

const (
	TestTemplate = "test.tmpl"
)

//go:embed "templates"
var FS embed.FS

type Mailer struct {
	Client *mailtrap.Client
	Config *Config
}

type UserInvitationPayload struct {
	Name  string
	Email string
}

func NewMailer(config Config) (*Mailer, error) {
	client, err := mailtrap.NewMailTrapClient(mailtrap.Config{
		FromPsw:   config.FromPsw,
		FromEmail: config.FromEmail,
	})
	if err != nil {
		return nil, err
	}

	mailer := Mailer{
		Client: client,
		Config: &config,
	}

	return &mailer, nil
}

func (o *Mailer) buildTemplate(templateFile string, data any) (*string, error) {
	tmpl, err := template.ParseFS(FS, "templates/"+templateFile)
	if err != nil {
		return nil, err
	}

	body := new(bytes.Buffer)
	if err = tmpl.ExecuteTemplate(body, "body", data); err != nil {
		return nil, err
	}

	str := body.String()
	return &str, nil
}

type TestPayload struct {
	Name string
}

func (o *Mailer) buildTestTemplate(payload TestPayload) (*string, error) {
	return o.buildTemplate(TestTemplate, payload)
}

func (o *Mailer) SendTestNotification(email, name string) error {
	body, err := o.buildTestTemplate(TestPayload{
		Name: name,
	})

	if err != nil {
		return err
	}

	recipient := email
	if o.Config.Env == "development" {
		recipient = o.Config.DevEmail
	}

	err = o.Client.Send(recipient, "Testing notification service", *body)
	if err != nil {
		return err
	}

	return nil
}
