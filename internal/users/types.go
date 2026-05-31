package users

import "fmt"

type Config struct {
	BaseURL *string `json:"baseURL"`
}

func (c Config) validate() error {
	if c.BaseURL == nil || *c.BaseURL == "" {
		return fmt.Errorf("base URL is required")
	}
	return nil
}

type User struct {
	UserID   string `json:"userId"`
	FCMToken string `json:"fcmToken"`
	Name     string `json:"name"`
	LastName string `json:"lastName"`
	Email    string `json:"email"`
}
