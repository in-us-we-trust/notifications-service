package users

import "fmt"

type Config struct {
	BaseURL *string `json:"baseURL"`
}

func (c Config) Validate() error {
	if c.BaseURL == nil || *c.BaseURL == "" {
		return fmt.Errorf("base URL is required")
	}
	return nil
}

type NotificationToken struct {
	AppDeviceID int    `json:"appDeviceId"`
	DeviceName  string `json:"deviceName"`
	Platform    string `json:"platform"`
	Token       string `json:"token"`
}

type User struct {
	ContactID          string              `json:"contactId"`
	UserID             string              `json:"userId"`
	Name               string              `json:"name"`
	LastName           string              `json:"lastName"`
	Email              string              `json:"email"`
	IsOwner            bool                `json:"isOwner"`
	NotificationTokens []NotificationToken `json:"notificationTokens"`
}

func Owners(contacts []User) []User {
	owners := make([]User, 0, len(contacts))
	for _, contact := range contacts {
		if contact.IsOwner {
			owners = append(owners, contact)
		}
	}
	return owners
}
