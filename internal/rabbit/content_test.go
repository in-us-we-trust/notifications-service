package rabbit

import "testing"

func TestIntruderContentUsesMessageAsIs(t *testing.T) {
	body := []byte(`{
		"severity": "high",
		"tags": ["danger", "doorbell"],
		"deviceId": "dev-1",
		"time": "2026-10-04T22:00:00Z",
		"title": "Peligro detectado",
		"message": "Se detectó una persona desconocida frente a la puerta.",
		"imageUrl": "https://example.test/foto.jpg"
	}`)

	content, err := ContentFrom(QueueIntruder, body)
	if err != nil {
		t.Fatalf("ContentFrom: %v", err)
	}

	if content.Title != "Peligro detectado" {
		t.Fatalf("title = %q", content.Title)
	}
	if content.Body != "Se detectó una persona desconocida frente a la puerta." {
		t.Fatalf("body = %q", content.Body)
	}
	if content.ImageURL != "https://example.test/foto.jpg" {
		t.Fatalf("image = %q", content.ImageURL)
	}
	if content.Data["type"] != QueueIntruder {
		t.Fatalf("type = %q", content.Data["type"])
	}
	if content.Data["deviceId"] != "dev-1" {
		t.Fatalf("deviceId = %q", content.Data["deviceId"])
	}
	if content.Data["imageUrl"] != "https://example.test/foto.jpg" {
		t.Fatalf("imageUrl = %q", content.Data["imageUrl"])
	}
	if content.Data["severity"] != "high" {
		t.Fatalf("severity = %q", content.Data["severity"])
	}
	if content.Data["tags"] != `["danger","doorbell"]` {
		t.Fatalf("tags = %q", content.Data["tags"])
	}
	if content.Data["time"] != "2026-10-04T22:00:00Z" {
		t.Fatalf("time = %q", content.Data["time"])
	}
	if content.PersonID != "" {
		t.Fatalf("personId = %q", content.PersonID)
	}
}

func TestUnknownPersonContent(t *testing.T) {
	body := []byte(`{
		"deviceId": "dev-1",
		"personId": "person-9",
		"time": "2026-10-04T22:00:00Z",
		"distinctDays": 4,
		"imageUrls": [
			"https://example.test/dia-1.jpg",
			"https://example.test/dia-3.jpg"
		]
	}`)

	content, err := ContentFrom(QueueUnknownPerson, body)
	if err != nil {
		t.Fatalf("ContentFrom: %v", err)
	}

	const wantBody = "Una persona apareció 4 días distintos. ¿Querés agregarla como conocida?"
	if content.Title != "Persona detectada" {
		t.Fatalf("title = %q", content.Title)
	}
	if content.Body != wantBody {
		t.Fatalf("body = %q", content.Body)
	}
	if content.ImageURL != "https://example.test/dia-1.jpg" {
		t.Fatalf("preview = %q", content.ImageURL)
	}
	if content.PersonID != "person-9" {
		t.Fatalf("personId = %q", content.PersonID)
	}
	if content.Data["type"] != QueueUnknownPerson {
		t.Fatalf("type = %q", content.Data["type"])
	}
	if content.Data["personId"] != "person-9" {
		t.Fatalf("data personId = %q", content.Data["personId"])
	}
	if content.Data["distinctDays"] != "4" {
		t.Fatalf("distinctDays = %q", content.Data["distinctDays"])
	}
	if content.Data["imageUrls"] != `["https://example.test/dia-1.jpg","https://example.test/dia-3.jpg"]` {
		t.Fatalf("imageUrls = %q", content.Data["imageUrls"])
	}
	if content.Data["message"] != wantBody {
		t.Fatalf("data message = %q", content.Data["message"])
	}
}

func TestUnknownPersonSingularDay(t *testing.T) {
	body := []byte(`{"deviceId":"dev-1","personId":"person-1","distinctDays":1}`)

	content, err := ContentFrom(QueueUnknownPerson, body)
	if err != nil {
		t.Fatalf("ContentFrom: %v", err)
	}

	const want = "Una persona apareció 1 día distinto. ¿Querés agregarla como conocida?"
	if content.Body != want {
		t.Fatalf("body = %q", content.Body)
	}
}

func TestUnknownPersonWithoutImagesStillBuildsPush(t *testing.T) {
	body := []byte(`{"deviceId":"dev-1","personId":"person-1","distinctDays":2}`)

	content, err := ContentFrom(QueueUnknownPerson, body)
	if err != nil {
		t.Fatalf("ContentFrom: %v", err)
	}
	if content.ImageURL != "" {
		t.Fatalf("image = %q", content.ImageURL)
	}
	if content.Data["imageUrls"] != "[]" {
		t.Fatalf("imageUrls = %q", content.Data["imageUrls"])
	}
	if _, ok := content.Data["time"]; ok {
		t.Fatalf("time should be omitted, got %q", content.Data["time"])
	}
}

func TestUnknownPersonRequiresPersonID(t *testing.T) {
	body := []byte(`{"deviceId":"dev-1","distinctDays":2}`)

	if _, err := ContentFrom(QueueUnknownPerson, body); err == nil {
		t.Fatal("expected error")
	}
}

func TestIntruderRequiresDeviceID(t *testing.T) {
	body := []byte(`{"title":"Peligro detectado","message":"hola"}`)

	if _, err := ContentFrom(QueueIntruder, body); err == nil {
		t.Fatal("expected error")
	}
}

func TestContentFromRejectsUnknownQueue(t *testing.T) {
	if _, err := ContentFrom("doorbell-notifications", []byte(`{}`)); err == nil {
		t.Fatal("expected error")
	}
}
