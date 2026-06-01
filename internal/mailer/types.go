package mailer

import "fmt"

type Config struct {
	FromEmail   string
	FromPsw     string
	FrontendUrl string
	DevEmail    string
	Env         string
}

func (c *Config) Validate() error {
	if c.FromEmail == "" {
		return fmt.Errorf("from email is required")
	}
	if c.FromPsw == "" {
		return fmt.Errorf("from password is required")
	}
	if c.FrontendUrl == "" {
		return fmt.Errorf("frontend url is required")
	}
	if c.DevEmail == "" {
		return fmt.Errorf("dev email is required")
	}
	if c.Env == "" {
		return fmt.Errorf("env is required")
	}
	return nil
}
