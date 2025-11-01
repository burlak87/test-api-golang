package service

import (
	"errors"
	"gosmol/internal/domain"
	"regexp"
	"time"

	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

type StudentsStorage interface {
	InsertStudents(students domain.Student) (int64, error)
	SelectStudents(email string) (domain.Student, error)
	RefreshStore(userID int64, token string, expiresAt time.Time) error
	RefreshGet(token string) (int64, error)
	RefreshDelete(token string) error
}


type Students struct {
	storage StudentsStorage
	jwtSecret  string
}

func NewStudents(storage StudentsStorage, jwt string) *Students{
	return &Students{storage: storage, jwtSecret: jwt}
}

func (s *Students) StudentsRegister(students domain.Student) error {
	if students.Firstname == "" || students.Lastname == "" || students.Email == "" {
		return errors.New("Invalid input")
	}

	if students.PasswordHash == "" || len(students.PasswordHash) < 8 {
		return errors.New("Invalid password input")
	}

	hasLetters, _ := regexp.MatchString(`[a-zA-Zа-яА-Я]`, students.PasswordHash)
	hasDigits, _ := regexp.MatchString(`[0-9]`, students.PasswordHash)
	hasSpecial, _ := regexp.MatchString(`[^a-zA-Zа-яА-Я0-9\s]`, students.PasswordHash)

	if hasLetters == false || hasDigits == false || hasSpecial == false {
		return errors.New("Invalid password input")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(students.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("Error hashing password")
	}

	student := domain.Student{Firstname: students.Firstname, Lastname: students.Lastname, Email: students.Email, PasswordHash: string(hash)}

	_, err = s.storage.InsertStudents(student)
	if err != nil {
		return err
	}

	return nil
}

func (s *Students) StudentsLogin(students domain.Student) (domain.TokenResponse,error) {
	if students.Email == "" {
		return domain.TokenResponse{}, errors.New("Invalid input")
	}

	stud, err := s.storage.SelectStudents(students.Email)
	if err != nil {
		return domain.TokenResponse{}, errors.New("Invalid credentials")
	}
	err = bcrypt.CompareHashAndPassword([]byte(stud.PasswordHash), []byte(students.PasswordHash))
	if err != nil {
		return domain.TokenResponse{}, errors.New("Invalid credentials")
	}

	accessToken, err := s.GenerateAccessToken(students.ID)
	if err != nil {
		return domain.TokenResponse{}, err
	}
	refreshToken, err := s.GenerateRefreshToken(students.ID)
	if err != nil {
		return domain.TokenResponse{}, err
	}

	return domain.TokenResponse{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func (s *Students) StudentsRefresh(refreshToken string) (domain.TokenResponse, error) {
	studentID, err := s.storage.RefreshGet(refreshToken)
	if err != nil {
		return  domain.TokenResponse{}, errors.New("Invalid refresh token")
	}

	accessToken, err := s.GenerateAccessToken(studentID)
	if err != nil {
		return domain.TokenResponse{}, err
	}
	newRefreshToken, err := s.GenerateRefreshToken(studentID)
	if err != nil {
		return domain.TokenResponse{}, err
	}
	s.storage.RefreshDelete(refreshToken)
	
	return domain.TokenResponse{AccessToken: accessToken, RefreshToken: newRefreshToken}, nil
}

func (s *Students) GenerateAccessToken(id int64) (string, error) {
	claims := jwt.MapClaims{
		"user_id": id,
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *Students) GenerateRefreshToken(id int64) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	
	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = id
	claims["exp"] = time.Now().Add(7 *24 * time.Hour).Unix()
	
	signed, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", err
	}
	
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	err = s.storage.RefreshStore(id, signed, expiresAt)
	return signed, err
}