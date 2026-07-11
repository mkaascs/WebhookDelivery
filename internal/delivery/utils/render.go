package utils

import (
	"errors"
	"fmt"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"
	"net/http"
	"webhook-delivery/internal/domain"
)

type Response struct {
	Errors []string `json:"errors,omitempty"`
}

func RenderError(w http.ResponseWriter, req *http.Request, statusCode int, errMessage string) {
	render.Status(req, statusCode)
	render.JSON(w, req, Response{Errors: []string{errMessage}})
}

func RenderValidationErrors(w http.ResponseWriter, req *http.Request, errs validator.ValidationErrors) {
	render.Status(req, http.StatusUnprocessableEntity)

	errMessages := make([]string, 0, len(errs))
	for _, err := range errs {
		switch err.ActualTag() {
		case "required":
			errMessages = append(errMessages, fmt.Sprintf("%s is required", err.Field()))
		case "url":
			errMessages = append(errMessages, fmt.Sprintf("%s is not a url", err.Field()))
		case "email":
			errMessages = append(errMessages, fmt.Sprintf("%s is not a valid email", err.Field()))
		case "min":
			errMessages = append(errMessages, fmt.Sprintf("%s must be greater or equal than %s", err.Field(), err.Param()))
		case "max":
			errMessages = append(errMessages, fmt.Sprintf("%s must be less or equal than %s", err.Field(), err.Param()))
		default:
			errMessages = append(errMessages, fmt.Sprintf("%s is invalid", err.Field()))
		}
	}

	render.JSON(w, req, Response{Errors: errMessages})
}

func TryRenderEndpointsError(w http.ResponseWriter, req *http.Request, err error) bool {
	if errors.Is(err, domain.ErrEndpointNotFound) {
		RenderError(w, req, http.StatusNotFound, "endpoint with this id not found")
		return true
	}

	if errors.Is(err, domain.ErrSubscriptionNotFound) {
		RenderError(w, req, http.StatusNotFound, "subscription with this id not found")
		return true
	}

	if errors.Is(err, domain.ErrSubscriptionAlreadyExists) {
		RenderError(w, req, http.StatusConflict, "subscription with this id is already exists")
		return true
	}

	return false
}
