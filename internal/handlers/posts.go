package handlers

import (
	"encoding/json"
	"forum/internal/models"
	"forum/internal/post_actions"
	"net/http"
	"strconv"
)

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
		writeJSONError(w, "Cant parse post id", http.StatusInternalServerError)
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

func (h *Handler) HandleDeletePost(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(userIDKey).(int)
	if !ok {
		writeJSONError(w, "User id not found", http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodDelete {
		writeJSONError(w, "Method is not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.PathValue("id")
	postID, err := strconv.Atoi(idStr)
	if err != nil {
		writeJSONError(w, "Cant parse post id", http.StatusInternalServerError)
		return
	}

	if err := post_actions.DeletePost(r.Context(), h.conn, userID, postID); err != nil {
		writeJSONError(w, "Cant delete the post", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) HandleGetPost(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		writeJSONError(w, "Method is not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.PathValue("id")

	postID, err := strconv.Atoi(idStr)
	if err != nil {
		writeJSONError(w, "Cant parse post id", http.StatusBadRequest)
		return
	}

	post, err := post_actions.GetPost(r.Context(), h.conn, postID)
	if err != nil {
		writeJSONError(w, "Cant get post", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, post)
}

func (h *Handler) HandleGetAllPosts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONError(w, "Method is not allowed", http.StatusMethodNotAllowed)
		return
	}

	posts, err := post_actions.GetAllPosts(r.Context(), h.conn)
	if err != nil {
		writeJSONError(w, "Cant get posts", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, posts)
}
