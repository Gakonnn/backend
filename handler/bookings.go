package handler

import (
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

type Favorites struct {
	ID       int    `db:"id"`
	UserID   int    `db:"user_id"`
	BookID   int    `db:"book_id"`
	BookName string `db:"book_name"`
}

func (h *Handler) addFavorite(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bookID := vars["id"]
	if bookID == "" {
		http.Error(rw, "Invalid URL", http.StatusInternalServerError)
		return
	}
	bookIDInt, err := strconv.Atoi(bookID)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	session, err := h.sess.Get(r, sessionName)
	if err != nil {
		http.Error(rw, "Unable to retrieve session", http.StatusInternalServerError)
		return
	}

	userID := session.Values["authUserID"]
	if userID == nil {
		http.Redirect(rw, r, "/login", http.StatusSeeOther)
		return
	}

	const checkFavorite = `
    SELECT COUNT(*) 
    FROM favorites 
    WHERE user_id = $1 AND book_id = $2
`
	var count int
	err = h.db.QueryRow(checkFavorite, userID, bookIDInt).Scan(&count)
	if err != nil {
		http.Error(rw, "Database error", http.StatusInternalServerError)
		return
	}

	if count > 0 {
		http.Redirect(rw, r, "/favorites", http.StatusFound)
		return
	}

	const insertFavorite = `INSERT INTO favorites(user_id, book_id) VALUES($1, $2)`
	_, err = h.db.Exec(insertFavorite, userID, bookIDInt)
	if err != nil {
		http.Error(rw, "Failed to add to favorites", http.StatusInternalServerError)
		return
	}

	http.Redirect(rw, r, "/favorites", http.StatusFound)
}

func (h *Handler) viewFavorites(rw http.ResponseWriter, r *http.Request) {
	session, err := h.sess.Get(r, sessionName)
	userID := session.Values["authUserID"]

	const selectFavorites = `
    SELECT f.id, f.user_id, f.book_id, b.book_name
    FROM favorites f
    JOIN books b ON f.book_id = b.id
    WHERE f.user_id = $1
`
	var favorites []Favorites
	err = h.db.Select(&favorites, selectFavorites, userID)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	// Отображение списка избранных книг
	if err := h.templates.ExecuteTemplate(rw, "my-bookings.html", favorites); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
}
func (h *Handler) removeFavorite(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bookID := vars["id"]
	if bookID == "" {
		http.Error(rw, "Invalid URL", http.StatusInternalServerError)
		return
	}
	bookIDInt, err := strconv.Atoi(bookID)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	session, err := h.sess.Get(r, sessionName)
	userID := session.Values["authUserID"]

	const deleteFavorite = `DELETE FROM favorites WHERE user_id = $1 AND book_id = $2`
	_, err = h.db.Exec(deleteFavorite, userID, bookIDInt)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(rw, r, "/favorites", http.StatusFound)
}
