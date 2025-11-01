package rest

import (
	"encoding/json"
	"gosmol/internal/apperror"
	"gosmol/internal/domain"
	"gosmol/pkg/logging"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type StudentsService interface {
	StudentsRegister(students domain.Student) error
	StudentsLogin(students domain.Student) (domain.TokenResponse, error)
	StudentsRefresh(token string) (domain.TokenResponse, error)
}

type StudentsHandler struct {
	service StudentsService
	logger *logging.Logger
}

func NewStudentsHandler(s StudentsService, l *logging.Logger) *StudentsHandler {
	return &StudentsHandler{
		service: s,
		logger: l,
	}
}

var stud []domain.Student

func (s *StudentsHandler) Register(router *httprouter.Router) {
	router.HandlerFunc(http.MethodPost, "api/auth/register", apperror.Middleware(s.signUp))
	router.HandlerFunc(http.MethodPost, "api/auth/login", apperror.Middleware(s.signIn))
	router.HandlerFunc(http.MethodPost, "api/auth/refresh", apperror.Middleware(s.refresh))
}

func (s *StudentsHandler) signUp(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Context-Type", "application/json")
	
	var student domain.Student
	if err := json.NewDecoder(r.Body).Decode(&student); err != nil {
		s.logger.Error("Failed to decode JSON: " + err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}
	defer r.Body.Close()

	if err := s.service.StudentsRegister(student); err != nil {
		s.logger.Error("Failed to register student: " + err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	json.NewEncoder(w).Encode(student)
	return  nil
}

func (s *StudentsHandler) signIn(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Context-Type", "application/json")
	
	var student domain.Student
	if err := json.NewDecoder(r.Body).Decode(&student); err != nil {
		s.logger.Error("Failed to decode JSON: " + err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}
	defer r.Body.Close()


	token, err := s.service.StudentsLogin(student)
	if err != nil {
		s.logger.Error("Failed to login student: " + err.Error())
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return err
	}

	json.NewEncoder(w).Encode(token)
	return  nil
}

func (s *StudentsHandler) refresh(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "application/json")
	var tokens struct{ RefreshToken string `json:"refresh_token"` }
	if err := json.NewDecoder(r.Body).Decode(&tokens); err != nil {
		s.logger.Error("Failed to decode JSON: " + err.Error())
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return err
	}
	defer r.Body.Close()

	token, err := s.service.StudentsRefresh(tokens.RefreshToken)
	if err != nil {
		s.logger.Error("Failed to refresh token: " + err.Error())
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return err
	}

	json.NewEncoder(w).Encode(token)
	return nil
}