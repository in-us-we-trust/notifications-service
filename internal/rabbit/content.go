package rabbit

import (
	"encoding/json"
	"fmt"
	"time"
)

const unknownPersonTitle = "Persona detectada"

func ContentFrom(queue string, body []byte) (PushContent, error) {
	switch queue {
	case QueueIntruder:
		var notification Notification
		if err := json.Unmarshal(body, &notification); err != nil {
			return PushContent{}, fmt.Errorf("invalid intruder notification: %w", err)
		}
		return notification.pushContent()
	case QueueUnknownPerson:
		var suggestion UnknownPersonSuggestion
		if err := json.Unmarshal(body, &suggestion); err != nil {
			return PushContent{}, fmt.Errorf("invalid unknown person notification: %w", err)
		}
		return suggestion.pushContent()
	default:
		return PushContent{}, fmt.Errorf("unsupported notification queue %q", queue)
	}
}

func (n Notification) pushContent() (PushContent, error) {
	if n.DeviceID == "" {
		return PushContent{}, fmt.Errorf("deviceId is required")
	}

	data := map[string]string{
		"type":     QueueIntruder,
		"deviceId": n.DeviceID,
		"title":    n.Title,
		"message":  n.Message,
	}
	putTime(data, n.Time)
	if n.Severity != "" {
		data["severity"] = n.Severity
	}
	if len(n.Tags) > 0 {
		encoded, err := json.Marshal(n.Tags)
		if err != nil {
			return PushContent{}, fmt.Errorf("encode tags: %w", err)
		}
		data["tags"] = string(encoded)
	}
	if n.ImageURL != "" {
		data["imageUrl"] = n.ImageURL
	}

	return PushContent{
		Type:     QueueIntruder,
		DeviceID: n.DeviceID,
		Title:    n.Title,
		Body:     n.Message,
		ImageURL: n.ImageURL,
		Data:     data,
	}, nil
}

func (s UnknownPersonSuggestion) pushContent() (PushContent, error) {
	if s.DeviceID == "" {
		return PushContent{}, fmt.Errorf("deviceId is required")
	}
	if s.PersonID == "" {
		return PushContent{}, fmt.Errorf("personId is required")
	}

	images := s.ImageURLs
	if images == nil {
		images = []string{}
	}
	encoded, err := json.Marshal(images)
	if err != nil {
		return PushContent{}, fmt.Errorf("encode imageUrls: %w", err)
	}

	body := unknownPersonBody(s.DistinctDays)
	data := map[string]string{
		"type":         QueueUnknownPerson,
		"deviceId":     s.DeviceID,
		"personId":     s.PersonID,
		"distinctDays": fmt.Sprintf("%d", s.DistinctDays),
		"imageUrls":    string(encoded),
		"title":        unknownPersonTitle,
		"message":      body,
	}
	putTime(data, s.Time)

	return PushContent{
		Type:     QueueUnknownPerson,
		DeviceID: s.DeviceID,
		PersonID: s.PersonID,
		Title:    unknownPersonTitle,
		Body:     body,
		ImageURL: firstImage(images),
		Data:     data,
	}, nil
}

func unknownPersonBody(days int) string {
	unit := "días distintos"
	if days == 1 {
		unit = "día distinto"
	}
	return fmt.Sprintf("Una persona apareció %d %s. ¿Querés agregarla como conocida?", days, unit)
}

func firstImage(urls []string) string {
	for _, url := range urls {
		if url != "" {
			return url
		}
	}
	return ""
}

func putTime(data map[string]string, at time.Time) {
	if at.IsZero() {
		return
	}
	data["time"] = at.UTC().Format(time.RFC3339)
}
