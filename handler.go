package cfb

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/smtp"
	"text/template"
)

var defaultTemplate = `
	<h1>Someone wants to connect with us!</h1>
	<ul>
		<li>Name: <b>{{.Name}}</b></li>
		<li>Email: <b>{{.Email}}</b></li>
	</ul>
	<h3>Message</h3>
	<p>
		{{.Message}}
	</p>
`

type Configuration struct {
	To []string

	FromEmail    string
	FromPassword string

	SMTPHost string
	SMTPPort string

	Subject  string
	Template string

	ErrorLogging bool
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
			errorResponse(w, err, "invalid request body", c)
		}

		var reqBody Request
		err = json.Unmarshal(reqBodyRaw, &reqBody)
		if err != nil {
			errorResponse(w, err, "invalid request body", c)
		}

		err = send(c, reqBody)
		if err != nil {
			errorResponse(w, err, "invalid request body", c)
		}
	}
}

func send(c Configuration, r Request) error {
	emailBody, err := parseTemplate(c, r)
	if err != nil {
		return errors.New("unable to parse email template")
	}

	var s = "Contact Form - " + r.Name
	if c.Subject != "" {
		s = c.Subject
	}
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	subject := "Subject: " + s + "!\n"
	msg := []byte(subject + mime + "\n" + emailBody)

	auth := smtp.PlainAuth("", c.FromEmail, c.FromPassword, c.SMTPHost)
	return smtp.SendMail(c.SMTPHost+":"+c.SMTPPort, auth, c.FromEmail, c.To, msg)
}

func parseTemplate(c Configuration, r Request) (string, error) {
	var templateUsed = defaultTemplate
	if c.Template != "" {
		templateUsed = c.Template
	}
	t, err := template.New("contact_form").Parse(templateUsed)
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

func errorResponse(w http.ResponseWriter, err error, message string, c Configuration) {
	if c.ErrorLogging {
		log.Print(err)
	}
	w.WriteHeader(http.StatusBadRequest)
	raw, _ := json.Marshal(Error{
		Message: message,
	})
	w.Write(raw)
}
