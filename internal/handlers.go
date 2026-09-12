package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"forum/internal/auth"
	"forum/internal/models"
	"forum/internal/post_actions"
	"log"
	"net/http"
	"strconv"
	"text/template"
	"time"

	"github.com/jackc/pgx/v5"
)

type contextKey string

const userIDKey contextKey = "userID"

type Handler struct {
	conn *pgx.Conn
}

func NewHandler(conn *pgx.Conn) *Handler {
	return &Handler{
		conn: conn,
	}
}

func writeJSONError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"message": message})
}

func MainHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		log.Println("Forbidden Method")
		writeJSONError(w, "Method is not Allowed", http.StatusMethodNotAllowed)
		return
	}

	tmp, err := template.ParseFiles("web/index.html")
	if err != nil {
		http.Error(w, "Failed to Parse the tmp", http.StatusInternalServerError)
		return
	}

	if err := tmp.Execute(w, nil); err != nil {
		http.Error(w, "Failed to Exec the tmp", http.StatusInternalServerError)
		return
	}

}

func (h *Handler) HandleRegistration(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case http.MethodGet:
		tmp, err := template.ParseFiles("web/register.html")
		if err != nil {
			writeJSONError(w, "Failed to parse tmp", http.StatusInternalServerError)
			return
		}

		if err := tmp.Execute(w, nil); err != nil {
			writeJSONError(w, "Failed to exec", http.StatusInternalServerError)
			return
		}
	case http.MethodPost:
		var user models.RegisterRequest

		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			writeJSONError(w, "Failed to decode the input", http.StatusBadRequest)
			return
		}

		_, err := auth.RegisterUser(r.Context(), h.conn, user)
		if err != nil {
			writeJSONError(w, "Cant register", http.StatusInternalServerError)
			log.Println(err)
			return
		}
		fmt.Println(user)

		w.WriteHeader(http.StatusCreated)
	default:
		writeJSONError(w, "Method is not Allowed", http.StatusMethodNotAllowed)
		return
	}

}

func (h *Handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		tmp, err := template.ParseFiles("web/login.html")
		if err != nil {
			writeJSONError(w, "Failed to parse tmp", http.StatusInternalServerError)
			return
		}

		if err := tmp.Execute(w, nil); err != nil {
			writeJSONError(w, "Failed to exec", http.StatusInternalServerError)
			return
		}
	case http.MethodPost:
		var user models.LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			writeJSONError(w, "cant decode json", http.StatusBadRequest)
			return
		}

		userid, valid := auth.AuthenticateUser(r.Context(), h.conn, user)
		if !valid {
			writeJSONError(w, "Wrong credentials", http.StatusUnauthorized)
			return
		}
		//create session
		access, err := auth.GenerateToken()
		if err != nil {
			writeJSONError(w, "cant gen access token", http.StatusInternalServerError)
			return
		}
		refresh, err := auth.GenerateToken()
		if err != nil {
			writeJSONError(w, "cant gen refresh token", http.StatusInternalServerError)
			return
		}
		access_hash := auth.HashToken(access)

		refresh_hash := auth.HashToken(refresh)

		if err := auth.InsertSession(r.Context(), h.conn, userid, access_hash, refresh_hash); err != nil {
			writeJSONError(w, "cant insert session", http.StatusInternalServerError)
			log.Println(err)
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:     "access_token",
			Value:    access,
			HttpOnly: true,
			Path:     "/",
			Expires:  time.Now().Add(15 * time.Minute),
		})
		http.SetCookie(w, &http.Cookie{
			Name:     "refresh_token",
			Value:    refresh,
			HttpOnly: true,
			Path:     "/refresh",
			Expires:  time.Now().Add(14 * 24 * time.Hour),
		})
		w.WriteHeader(http.StatusOK)
	default:
		writeJSONError(w, "Method is not allowed", http.StatusMethodNotAllowed)
		return
	}
}

func (h *Handler) HandleRefresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		writeJSONError(w, "no cookie", http.StatusUnauthorized)
		return
	}

	userID, valid := auth.ValidateToken(r.Context(), h.conn, cookie.Value, "refresh")

	if !valid {
		writeJSONError(w, "invalid or expired refresh token", http.StatusUnauthorized)
		return
	}

	new_access_token, err := auth.GenerateToken()
	if err != nil {
		writeJSONError(w, "cant gen new access token", http.StatusInternalServerError)
		return
	}
	new_access_token_hash := auth.HashToken(new_access_token)

	new_refresh_token, err := auth.GenerateToken()
	if err != nil {
		writeJSONError(w, "cant gen new refresh token", http.StatusInternalServerError)
		return
	}
	new_refresh_token_hash := auth.HashToken(new_refresh_token)
	old_refresh_token_hash := auth.HashToken(cookie.Value)
	if !auth.UpdateTokens(r.Context(), h.conn, userID, new_access_token_hash, new_refresh_token_hash, old_refresh_token_hash) {
		writeJSONError(w, "cant update access token", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    new_access_token,
		HttpOnly: true,
		Path:     "/",
		Expires:  time.Now().Add(15 * time.Minute),
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    new_refresh_token,
		HttpOnly: true,
		Path:     "/refresh",
		Expires:  time.Now().Add(14 * 24 * time.Hour),
	})

	w.WriteHeader(http.StatusOK)

}

func (h *Handler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, "Method is not allowed", http.StatusMethodNotAllowed)
		return
	}

	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		writeJSONError(w, "Refresh token not available", http.StatusUnauthorized)
		return
	}

	refresh_hash := auth.HashToken(cookie.Value)

	if err := auth.DeleteSession(r.Context(), h.conn, refresh_hash); err != nil {
		writeJSONError(w, "cant delete session", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		HttpOnly: true,
		Path:     "/refresh",
		MaxAge:   -1,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		HttpOnly: true,
		Path:     "/",
		MaxAge:   -1,
	})

	w.WriteHeader(http.StatusOK)

}

func (h *Handler) Middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("access_token")
		if err != nil {
			writeJSONError(w, "Access token not available", http.StatusUnauthorized)
			return
		}

		userID, valid := auth.ValidateToken(r.Context(), h.conn, cookie.Value, "access")
		if !valid {
			writeJSONError(w, "invalid or expired access token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next(w, r.WithContext(ctx))
	}
}

func (h *Handler) HandleCreatePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, "Method is not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value(userIDKey).(int)
	if !ok {
		writeJSONError(w, "User id not found", http.StatusUnauthorized)
		return
	}

	var post models.PostRequest

	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		writeJSONError(w, "Failed to decode the input", http.StatusBadRequest)
		return
	}

	if err := post_actions.InsertPost(r.Context(), h.conn, userID, post); err != nil {
		writeJSONError(w, "Failed to insert post to a DB", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)

}

func (h *Handler) HandleEditPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		writeJSONError(w, "Method is not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.PathValue("id")

	postID, err := strconv.Atoi(idStr)
	if err != nil {
		writeJSONError(w, "Cant find parse post id", http.StatusInternalServerError)
		return
	}

	userID, ok := r.Context().Value(userIDKey).(int)
	if !ok {
		writeJSONError(w, "User id not found", http.StatusUnauthorized)
		return
	}

	var post models.PostRequest

	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		writeJSONError(w, "Failed to decode the input", http.StatusBadRequest)
		return
	}

	if err := post_actions.EditPost(r.Context(), h.conn, userID, postID, post); err != nil {
		writeJSONError(w, "Cant update table", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

}
