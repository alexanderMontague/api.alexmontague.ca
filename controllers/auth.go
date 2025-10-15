package controllers

import (
	"encoding/json"
	"log"
	"net/http"

	"api.alexmontague.ca/helpers"
	"api.alexmontague.ca/internal/database/repository"
	"api.alexmontague.ca/middleware"
)

func Register(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	var req repository.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logAndRespondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if req.Email == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(helpers.Response{Error: true, Code: 400, Message: "Email and password are required"})
		return
	}

	existingUser, _ := repository.GetUserByEmail(req.Email)
	if existingUser != nil {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(helpers.Response{Error: true, Code: 409, Message: "User already exists"})
		return
	}

	user, err := repository.CreateUser(req.Email, req.Password)
	if err != nil {
		logAndRespondError(w, http.StatusInternalServerError, "Failed to create user", err)
		return
	}

	if err := repository.CreateDefaultCategory(user.ID); err != nil {
		log.Printf("[Warning] Failed to create default category for user %d: %v", user.ID, err)
	}

	token, err := helpers.GenerateToken(user.ID, user.Email)
	if err != nil {
		logAndRespondError(w, http.StatusInternalServerError, "Failed to generate token", err)
		return
	}

	response := map[string]interface{}{
		"user": map[string]interface{}{
			"id":    user.ID,
			"email": user.Email,
		},
		"token": token,
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	var req repository.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logAndRespondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if req.Email == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(helpers.Response{Error: true, Code: 400, Message: "Email and password are required"})
		return
	}

	user, err := repository.GetUserByEmail(req.Email)
	if err != nil {
		log.Printf("[Auth Error] Login attempt for non-existent user: %s", req.Email)
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(helpers.Response{Error: true, Code: 401, Message: "Invalid credentials"})
		return
	}

	if !repository.ValidatePassword(user, req.Password) {
		log.Printf("[Auth Error] Invalid password attempt for user: %s", req.Email)
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(helpers.Response{Error: true, Code: 401, Message: "Invalid credentials"})
		return
	}

	token, err := helpers.GenerateToken(user.ID, user.Email)
	if err != nil {
		logAndRespondError(w, http.StatusInternalServerError, "Failed to generate token", err)
		return
	}

	response := map[string]interface{}{
		"user": map[string]interface{}{
			"id":    user.ID,
			"email": user.Email,
		},
		"token": token,
	}

	json.NewEncoder(w).Encode(response)
}

func GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	authUser := middleware.GetAuthUser(r)
	if authUser == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(helpers.Response{Error: true, Code: 401, Message: "Unauthorized"})
		return
	}

	user, err := repository.GetUserByID(authUser.UserID)
	if err != nil {
		logAndRespondError(w, http.StatusInternalServerError, "Failed to fetch user", err)
		return
	}

	response := map[string]interface{}{
		"id":    user.ID,
		"email": user.Email,
	}

	json.NewEncoder(w).Encode(response)
}
