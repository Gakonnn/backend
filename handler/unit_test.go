package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/stretchr/testify/require"
)

const sessionName = "library-session"

type mockSession struct {
	session *sessions.Session
}

func (m *mockSession) Get(r *http.Request, name string) (*sessions.Session, error) {
	if m.session != nil {
		return m.session, nil
	}
	return nil, errors.New("session not found")
}

type MockHandler struct {
	sess *mockSession
}

func (h *MockHandler) addFavorite(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bookID := vars["id"]
	if bookID == "" {
		http.Error(rw, "Invalid URL", http.StatusBadRequest)
		return
	}

	_, err := strconv.Atoi(bookID)
	if err != nil {
		http.Error(rw, "Invalid book ID", http.StatusBadRequest)
		return
	}

	session, err := h.sess.Get(r, sessionName)
	if err != nil {
		http.Error(rw, "Session error", http.StatusUnauthorized)
		return
	}

	userID := session.Values["authUserID"]
	if userID == nil {
		http.Redirect(rw, r, "/login", http.StatusSeeOther)
		return
	}

	rw.Header().Set("Location", "/favorites")
	rw.WriteHeader(http.StatusFound)
}

func TestAddFavorite_NoDB(t *testing.T) {
	session := sessions.NewSession(nil, "test-session")
	session.Values["authUserID"] = 1

	h := &MockHandler{
		sess: &mockSession{
			session: session,
		},
	}

	req := httptest.NewRequest("POST", "/add/123", nil)
	req = mux.SetURLVars(req, map[string]string{"id": strconv.Itoa(123)})

	rr := httptest.NewRecorder()

	h.addFavorite(rr, req)

	require.Equal(t, http.StatusFound, rr.Code)
	require.Equal(t, "/favorites", rr.Header().Get("Location"))
}
