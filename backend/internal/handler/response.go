package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-playground/validator/v10"
	"kopirise/backend/internal/models"
)

var validate = validator.New()

// RespondJSON menulis payload JSON standar.
func RespondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// RespondSuccess mengirim respon sukses dengan data.
func RespondSuccess(w http.ResponseWriter, data interface{}) {
	RespondJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    data,
	})
}

// RespondCreated mengirim respon sukses pembuatan data baru (201 Created).
func RespondCreated(w http.ResponseWriter, data interface{}) {
	RespondJSON(w, http.StatusCreated, models.APIResponse{
		Success: true,
		Data:    data,
	})
}

// RespondPagination mengirim respon berpaginasi standar.
func RespondPagination(w http.ResponseWriter, data interface{}, meta *models.PaginationMeta) {
	RespondJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    data,
		Meta:    meta,
	})
}

// RespondError mengirim respon error standar JSON.
func RespondError(w http.ResponseWriter, status int, code, message string, details interface{}) {
	RespondJSON(w, status, models.APIResponse{
		Success: false,
		Error: &models.APIErrorDetail{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}

// DecodeAndValidate menguraikan JSON request body dan memvalidasi tag validator v10.
func DecodeAndValidate(r *http.Request, target interface{}) error {
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		return errors.New("format payload JSON tidak valid atau malformed")
	}
	if err := validate.Struct(target); err != nil {
		return err
	}
	return nil
}
