package handler

import (
	"encoding/json"
	"fmt"
	"Practice5/internal/usecase"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type UserHandler struct {
	usecase *usecase.UserUsecase
}

func NewUserHandler(uc *usecase.UserUsecase) *UserHandler {
	return &UserHandler{
		usecase: uc,
	}
}

func (h *UserHandler) GetPaginatedUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	pageStr := r.URL.Query().Get("page")
	if pageStr == "" {
		pageStr = "1"
	}
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSizeStr := r.URL.Query().Get("pageSize")
	if pageSizeStr == "" {
		pageSizeStr = "10"
	}
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		pageSize = 10
	}

	orderBy := r.URL.Query().Get("orderBy")
	if orderBy == "" {
		orderBy = "id"
	}

	filters := make(map[string]interface{})

	if id := r.URL.Query().Get("id"); id != "" {
		idInt, err := strconv.Atoi(id)
		if err == nil {
			filters["id"] = idInt
		}
	}

	if name := r.URL.Query().Get("name"); name != "" {
		filters["name"] = name
	}

	if email := r.URL.Query().Get("email"); email != "" {
		filters["email"] = email
	}

	if gender := r.URL.Query().Get("gender"); gender != "" {
		filters["gender"] = gender
	}

	if birthDate := r.URL.Query().Get("birth_date"); birthDate != "" {
		filters["birth_date"] = birthDate
	}

	result, err := h.usecase.GetPaginatedUsers(page, pageSize, filters, orderBy)
	if err != nil {
		log.Printf("Error fetching users: %v", err)
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

func (h *UserHandler) GetCommonFriends(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	userID, err := h.extractID(r.URL.Path, "/users/", "/common-friends")
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	friendIDStr := r.URL.Query().Get("with")
	if friendIDStr == "" {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error":"missing 'with' parameter"}`, http.StatusBadRequest)
		return
	}

	friendID, err := strconv.Atoi(friendIDStr)
	if err != nil || friendID <= 0 {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error":"invalid friend ID"}`, http.StatusBadRequest)
		return
	}

	commonFriends, err := h.usecase.GetCommonFriends(userID, friendID)
	if err != nil {
		log.Printf("Error fetching common friends: %v", err)
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"user_id":         userID,
		"friend_id":       friendID,
		"common_friends":  commonFriends,
		"count":           len(commonFriends),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *UserHandler) GetFriendsOfUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	userID, err := h.extractID(r.URL.Path, "/users/", "/friends")
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	friends, err := h.usecase.GetFriendsOfUser(userID)
	if err != nil {
		log.Printf("Error fetching friends: %v", err)
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"user_id": userID,
		"friends": friends,
		"count":   len(friends),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *UserHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "healthy",
		"message": "API is running",
	})
}

func (h *UserHandler) extractID(path string, prefix string, suffix string) (int, error) {
	path = strings.TrimPrefix(path, prefix)
	path = strings.TrimSuffix(path, suffix)
	path = strings.TrimSuffix(path, "/")

	if path == "" {
		return 0, fmt.Errorf("invalid user ID")
	}

	id, err := strconv.Atoi(path)
	if err != nil {
		return 0, fmt.Errorf("invalid user ID: must be a number")
	}

	if id <= 0 {
		return 0, fmt.Errorf("invalid user ID: must be positive")
	}

	return id, nil
}