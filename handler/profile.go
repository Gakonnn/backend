package handler

import (
	"log"
	"net/http"
)

func (h *Handler) viewProfile(rw http.ResponseWriter, r *http.Request) {
	session, err := h.sess.Get(r, sessionName)
	if err != nil {
		http.Error(rw, "Unable to retrieve session", http.StatusInternalServerError)
		return
	}

	userID, ok := session.Values["authUserID"].(int)
	if !ok || userID == 0 {
		http.Redirect(rw, r, "/login", http.StatusSeeOther)
		return
	}

	const selectUser = `
    SELECT id, first_name, last_name ,email FROM users WHERE id = $1
`
	var user struct {
		ID       int    `db:"id"`
		Fistname string `db:"first_name"`
		Lastname string `db:"last_name"`
		Email    string `db:"email"`
	}
	log.Println("UserID from session:", userID)

	err = h.db.Get(&user, selectUser, userID)
	if err != nil {
		http.Error(rw, "User not found", http.StatusInternalServerError)
		return
	}

	if err := h.templates.ExecuteTemplate(rw, "profile.html", user); err != nil {
		http.Error(rw, "Template rendering failed", http.StatusInternalServerError)
	}
}
func (h *Handler) updateProfile(rw http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(rw, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	session, err := h.sess.Get(r, sessionName)
	if err != nil {
		http.Error(rw, "Unable to retrieve session", http.StatusInternalServerError)
		return
	}

	userID, ok := session.Values["authUserID"].(int)
	if !ok || userID == 0 {
		http.Redirect(rw, r, "/login", http.StatusSeeOther)
		return
	}

	fullName := r.FormValue("first_name")
	lastName := r.FormValue("last_name")
	email := r.FormValue("email")

	const updateUser = `
    UPDATE users
    SET first_name = $1, last_name = $2, email = $3
    WHERE id = $4
`
	_, err = h.db.Exec(updateUser, fullName, lastName, email, userID)
	if err != nil {
		http.Error(rw, "Failed to update profile", http.StatusInternalServerError)
		return
	}

	http.Redirect(rw, r, "/profile", http.StatusSeeOther)
}
