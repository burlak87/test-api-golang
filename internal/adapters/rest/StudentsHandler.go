package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	
	"gosmol/internal/apperror"
	"gosmol/internal/domain"
	"gosmol/pkg/logging"
	
	"github.com/julienschmidt/httprouter"
)

type StudentsService interface {
	StudentsRegister(students domain.Student) (domain.Student, error)
	StudentsLogin(students domain.Student) (domain.TokenResponse, domain.TwoFaCodes, error)
	StudentsRefresh(token string) (domain.TokenResponse, error)
	StudentsSendEmailCode(tempToken string) error
    VerifyCode(code domain.Code) (domain.TokenResponse, error)
    EnableTwoFA(userID int64) error
    DisableTwoFA(userID int64, password string) error
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

func (s *StudentsHandler) Register(router *httprouter.Router, jwtSecret string) {
	router.HandlerFunc(http.MethodPost, "/api/auth/register", apperror.Middleware(s.signUp))
	router.HandlerFunc(http.MethodPost, "/api/auth/login", apperror.Middleware(s.signIn))
	router.HandlerFunc(http.MethodPost, "/api/auth/refresh", apperror.Middleware(s.refresh))
	router.HandlerFunc(http.MethodPost, "/api/auth/send-code", apperror.Middleware(s.sendEmailToken))
	router.HandlerFunc(http.MethodPost, "/api/auth/verify-code", apperror.Middleware(s.verifyCode))
	router.Handler(http.MethodPost, "/api/auth/enable-2fa", apperror.JWTMiddleware(jwtSecret, http.HandlerFunc(apperror.Middleware(s.enableTwoFA))))
	router.Handler(http.MethodPost, "/api/auth/disable-2fa", apperror.JWTMiddleware(jwtSecret, http.HandlerFunc(apperror.Middleware(s.disableTwoFA))))
}

func (s *StudentsHandler) signUp(w http.ResponseWriter, r *http.Request) error {
    w.Header().Set("Content-Type", "application/json")
    
    var student domain.Student
    if err := json.NewDecoder(r.Body).Decode(&student); err != nil {
        s.logger.Error("Failed to decode JSON: " + err.Error())
        http.Error(w, err.Error(), http.StatusBadRequest)
        return err
    }
    defer r.Body.Close()

    fmt.Printf("DEBUG HANDLER REGISTER: Received - Firstname: %s, Email: %s\n", 
        student.Firstname, student.Email)

    createdStudent, err := s.service.StudentsRegister(student)
    if err != nil {
        s.logger.Error("Failed to register student: " + err.Error())
        http.Error(w, err.Error(), http.StatusBadRequest)
        return err
    }

    fmt.Printf("DEBUG HANDLER REGISTER: Student created with ID: %d\n", createdStudent.ID)
    
    w.WriteHeader(http.StatusCreated)
    return json.NewEncoder(w).Encode(createdStudent)
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


	accessToken, tempToken, err := s.service.StudentsLogin(student)
	if err != nil {
		s.logger.Error("Failed to login student: " + err.Error())
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return err
	}

	if tempToken.RequiresTwoFa {
		return json.NewEncoder(w).Encode(tempToken)
	}
	
	json.NewEncoder(w).Encode(accessToken)
	return nil
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

func (s *StudentsHandler) sendEmailToken(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "application/json")
	var req struct { TempToken string `json:"temp_token"` }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.logger.Error("Failed to decode JSON: " + err.Error())
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return err
	}
	defer r.Body.Close()
	
	err := s.service.StudentsSendEmailCode(req.TempToken)
	if err != nil {
		s.logger.Error("Failed to temp token: " + err.Error())
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return err
	}
	
	res := map[string]bool{"success": true}
	json.NewEncoder(w).Encode(res)
	return nil
}

func (s *StudentsHandler) verifyCode(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "application/json")
	var code domain.Code
	
	if err := json.NewDecoder(r.Body).Decode(&code); err != nil {
		s.logger.Error("Failed to decode JSON: " + err.Error())
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return err
	}
	defer r.Body.Close()
	
	tokenRes, err := s.service.VerifyCode(code)
	if err != nil {
		s.logger.Error("Failed to verify code: " + err.Error())
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return err
	}
	
	json.NewEncoder(w).Encode(tokenRes)
	return nil
}

func (s *StudentsHandler) enableTwoFA(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "application/json")
	
	userID, ok := r.Context().Value("studentID").(int64)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return nil
	}
	
	err := s.service.EnableTwoFA(userID)
	if err != nil {
		s.logger.Error("Failed to enable 2FA: " + err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}
	
	res := map[string]bool{"success": true}
	json.NewEncoder(w).Encode(res)
	return nil
}

func (s *StudentsHandler) disableTwoFA(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "application/json")
	
	userID, ok := r.Context().Value("studentID").(int64)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return nil
	}
	
	var req domain.TwoFaToggleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.logger.Error("Failed to decode JSON: " + err.Error())
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return err
	}
	defer r.Body.Close()
	
	err := s.service.DisableTwoFA(userID, req.Password)
	if err != nil {
		s.logger.Error("Failed to disable 2FA: " + err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	res := map[string]bool{"success": true}
	json.NewEncoder(w).Encode(res)
	return nil
}