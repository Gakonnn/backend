package handler

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"html/template"
	"net/http"
	"net/smtp"

	"crypto/rand"
	validation "github.com/go-ozzo/ozzo-validation"
	"golang.org/x/crypto/bcrypt"
)

type SignUp struct {
	ID                int    `db:"id"`
	FirstName         string `db:"first_name"`
	LastName          string `db:"last_name"`
	Email             string `db:"email"`
	Password          string `db:"password"`
	ConfirmPassword   string
	IsVerified        bool   `db:"is_verified"`
	VerificationToken string `db:"verification_token"`
}

type SignUpForm struct {
	SingUp SignUp
	Errors map[string]string
}

func GenerateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func (s *SignUp) Validate() error {
	return validation.ValidateStruct(s,
		validation.Field(&s.FirstName,
			validation.Required.Error("This field is must required")),
		validation.Field(&s.LastName,
			validation.Required.Error("This field is must required")),
		validation.Field(&s.Email,
			validation.Required.Error("This field is must required")),
		validation.Field(&s.Password,
			validation.Required.Error("This field is must required")),
		validation.Field(&s.ConfirmPassword,
			validation.Required.Error("This field is must required")))
}

func (h *Handler) signUp(rw http.ResponseWriter, r *http.Request) {
	vErrs := map[string]string{}
	signup := SignUp{}
	h.loadSignUpForm(rw, signup, vErrs)
}

func (h *Handler) signUpCheck(rw http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	var signup SignUp
	if err := h.decoder.Decode(&signup, r.PostForm); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	if signup.Password != signup.ConfirmPassword {
		formData := SignUpForm{
			SingUp: signup,
			Errors: map[string]string{"Password": "The password does not match with the confirm password"},
		}
		if err := h.templates.ExecuteTemplate(rw, "signup.html", formData); err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}

	if err := signup.Validate(); err != nil {
		vErrors, ok := err.(validation.Errors)
		if ok {
			vErrs := make(map[string]string)
			for key, value := range vErrors {
				vErrs[key] = value.Error()
			}
			h.loadSignUpForm(rw, signup, vErrs)
			return
		}
	}

	// Генерация токена верификации
	verificationToken, err := GenerateToken()
	if err != nil {
		http.Error(rw, "Error generating verification token", http.StatusInternalServerError)
		return
	}

	// Хэширование пароля
	pass, err := bcrypt.GenerateFromPassword([]byte(signup.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(rw, "Error hashing password", http.StatusInternalServerError)
		return
	}

	// Сохранение пользователя с токеном
	const insertUserQuery = `
		INSERT INTO users (first_name, last_name, email, password, is_verified, verification_token)
		VALUES ($1, $2, $3, $4, false, $5)`
	_, err = h.db.Exec(insertUserQuery, signup.FirstName, signup.LastName, signup.Email, string(pass), verificationToken)
	if err != nil {
		http.Error(rw, "Error saving user", http.StatusInternalServerError)
		return
	}

	// Создание ссылки для подтверждения
	verifyURL := fmt.Sprintf("http://localhost:3000/verify-email?token=%s", verificationToken)

	// Отправка письма
	err = sendVerificationEmail(signup.Email, signup.FirstName, verifyURL)
	if err != nil {
		http.Error(rw, "Error sending email", http.StatusInternalServerError)
		return
	}

	// Перенаправление на страницу входа
	http.Redirect(rw, r, "/login", http.StatusTemporaryRedirect)
}
func sendVerificationEmail(email, name, link string) error {
	from := "e_book_aitu@zohomail.com"
	password := "gakon2006"
	to := []string{email}
	smtpHost := "smtp.zoho.com"
	smtpPort := "587"

	auth := smtp.PlainAuth("", from, password, smtpHost)
	tmpl, err := template.ParseFiles("templates/mail-template.html")
	if err != nil {
		return err
	}

	var body bytes.Buffer
	body.Write([]byte(fmt.Sprintf("Subject: Verification Mail\nMIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n")))
	err = tmpl.Execute(&body, struct {
		Name string
		Link string
	}{
		Name: name,
		Link: link,
	})
	if err != nil {
		return err
	}

	return smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, body.Bytes())
}
func (h *Handler) verifyEmail(rw http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(rw, "Invalid verification token", http.StatusBadRequest)
		return
	}

	const verifyUserQuery = `
		UPDATE users SET is_verified = true, verification_token = NULL
		WHERE verification_token = $1`
	res, err := h.db.Exec(verifyUserQuery, token)
	if err != nil {
		http.Error(rw, "Error verifying email", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil || rowsAffected == 0 {
		http.Error(rw, "Invalid or expired token", http.StatusBadRequest)
		return
	}

	http.Redirect(rw, r, "/login", http.StatusSeeOther)
}

func (h *Handler) loadSignUpForm(rw http.ResponseWriter, singup SignUp, errs map[string]string) {
	data := SignUpForm{
		SingUp: singup,
		Errors: errs,
	}
	if err := h.templates.ExecuteTemplate(rw, "signup.html", data); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
}
