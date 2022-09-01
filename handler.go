package cfb

import (
	"bytes"
	"encoding/json"
	"errors"
	"html/template"
	"io"
	"net/http"
	"net/smtp"
)

var defaultTemplate = `
	Name: {{.Name}}
	Email: {{.Email}}
	Message: {{.Message}}
`

type Configuration struct {
	To []string

	FromEmail    string
	FromPassword string

	SMTPHost string
	SMTPPort string

	Subject string
}

type Request struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Message string `json:"message"`
}

type Error struct {
	Message string `json:"message"`
}

func Handler(c Configuration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqBodyRaw, err := io.ReadAll(r.Body)
		if err != nil {
			errorResponse(w, "invalid request body")
		}

		var reqBody Request
		err = json.Unmarshal(reqBodyRaw, &reqBody)
		if err != nil {
			errorResponse(w, "invalid request body")
		}

		send(c, reqBody)
	}
}

func send(c Configuration, r Request) error {
	emailBody, err := parseTemplate(r)
	if err != nil {
		return errors.New("unable to parse email template")
	}

	mime := "MIME-version: 1.0;\nContent-Type: text/plain; charset=\"UTF-8\";\n\n"
	subject := "Subject: " + "Contact Form" + "!\n"
	msg := []byte(subject + mime + "\n" + emailBody)

	auth := smtp.PlainAuth("", c.FromEmail, c.FromPassword, c.SMTPHost)
	return smtp.SendMail(c.SMTPHost+":"+c.SMTPPort, auth, c.FromEmail, c.To, msg)
}

func parseTemplate(r Request) (string, error) {
	t, err := template.New("contact_form").Parse(defaultTemplate)
	if err != nil {
		return "", err
	}
	buf := new(bytes.Buffer)
	if err = t.Execute(buf, r); err != nil {
		return "", err
	}
	body := buf.String()
	return body, nil
}

func errorResponse(w http.ResponseWriter, message string) {
	w.WriteHeader(http.StatusBadRequest)
	raw, _ := json.Marshal(Error{
		Message: message,
	})
	w.Write(raw)
}
