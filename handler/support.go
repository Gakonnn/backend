package handler

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/smtp"
)

func (h *Handler) viewSupport(rw http.ResponseWriter, r *http.Request) {
	h.templates.ExecuteTemplate(rw, "support.html", nil)
}

func (h *Handler) sendSupportMessage(rw http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(rw, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	name := r.PostFormValue("name")
	email := r.PostFormValue("email")
	message := r.PostFormValue("message")

	// Handle file upload
	file, fileHeader, err := r.FormFile("attachment")
	if err != nil {
		log.Println("Error retrieving file:", err)
		http.Error(rw, "Failed to retrieve file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Read the file content
	fileBuffer := &bytes.Buffer{}
	_, err = io.Copy(fileBuffer, file)
	if err != nil {
		log.Println("Error reading file:", err)
		http.Error(rw, "Failed to process file", http.StatusInternalServerError)
		return
	}

	body := fmt.Sprintf("Message from: %s (%s)\n\nMessage:\n%s", name, email, message)

	smtpHost := "smtp.zoho.com"
	smtpPort := "587"
	smtpUser := "e_book_aitu@zohomail.com"
	smtpPass := "gakon2006"

	from := smtpUser
	to := []string{"galimjantugelbaev@gmail.com"}

	// Create email with attachment
	emailBuffer := &bytes.Buffer{}
	writer := multipart.NewWriter(emailBuffer)

	// Add email headers
	headers := map[string]string{
		"From":    from,
		"To":      to[0],
		"Subject": "New support request",
	}
	for k, v := range headers {
		emailBuffer.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	emailBuffer.WriteString("Content-Type: multipart/mixed; boundary=" + writer.Boundary() + "\r\n\r\n")

	// Add the plain text message
	textPart, err := writer.CreatePart(map[string][]string{"Content-Type": {"text/plain; charset=utf-8"}})
	if err != nil {
		log.Println("Error creating text part:", err)
		http.Error(rw, "Failed to create email body", http.StatusInternalServerError)
		return
	}
	textPart.Write([]byte(body))

	// Add the file attachment
	filePart, err := writer.CreateFormFile("attachment", fileHeader.Filename)
	if err != nil {
		log.Println("Error creating file part:", err)
		http.Error(rw, "Failed to attach file", http.StatusInternalServerError)
		return
	}
	filePart.Write(fileBuffer.Bytes())

	writer.Close()

	// Send the email
	err = smtp.SendMail(smtpHost+":"+smtpPort, smtp.PlainAuth("", smtpUser, smtpPass, smtpHost), from, to, emailBuffer.Bytes())
	if err != nil {
		log.Println("Error sending email:", err)
		http.Error(rw, "Failed to send message", http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(rw, "Your message has been sent. Thank you!")
}
