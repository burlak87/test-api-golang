package service

import (
	"errors"
	"fmt"
	"gosmol/internal/domain"
	"math"
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
	StudentBlocked(email string, windowStart time.Time) ([]map[string]interface{}, error)
	LogAttempt(email string, result bool, attemptTime time.Time) error
	GetFailedLogAttempts(email string, windowStart time.Time) (int, error)
	BlockStudent(email, blockedUntil string) error
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
		return domain.TokenResponse{}, errors.New("invalid input")
	}
	
	blocked, minutesLeft, err := s.IsUserBlocked(students.Email)
	if err != nil {
		return domain.TokenResponse{}, err
	}
	
	if blocked {
		return domain.TokenResponse{}, fmt.Errorf("your account is blocked for %d minutes", minutesLeft)
	}
	
	stud, err := s.storage.SelectStudents(students.Email)
	if err != nil {
		s.LogLoginAttempt(students.Email, false) // Логируем неудачную попытку
		return domain.TokenResponse{}, errors.New("invalid credentials")
	}
	
	err = bcrypt.CompareHashAndPassword([]byte(stud.PasswordHash), []byte(students.PasswordHash))
	if err != nil {
		s.LogLoginAttempt(students.Email, false) // Логируем неудачную попытку
		return domain.TokenResponse{}, errors.New("invalid credentials")
	}
	
	attempts, err := s.GetFailedAttempts(students.Email)
	if err != nil {
		return domain.TokenResponse{}, err
	}
	
	maxAttempts := int64(5)
	if attempts >= maxAttempts {
		s.BlockUser(students.Email)
		return domain.TokenResponse{}, errors.New("too many failed attempts, account blocked")
	}

	accessToken, err := s.GenerateAccessToken(stud.ID)
	if err != nil {
		return domain.TokenResponse{}, err
	}
	
	refreshToken, err := s.GenerateRefreshToken(stud.ID)
	if err != nil {
		return domain.TokenResponse{}, err
	}
	
	s.LogLoginAttempt(students.Email, true) // Логируем успешную попытку
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

func (s *Students) IsUserBlocked(email string) (bool, int64, error) {
	now := time.Now().UTC()
	windowStart := now
	
	result, err := s.storage.StudentBlocked(email, windowStart)
	if err != nil {
		fmt.Printf("Ошибка проверки блокировки: %v\n", err)
		return false, 0, err
	}
	
	if len(result) > 0 {
		blockedUntilStr, ok := result[0]["blocked_until"].(string)
		if !ok {
			return false, 0, errors.New("invalid format for blocked_until")
		}
	
		blockedUntil, err := time.Parse(time.RFC3339, blockedUntilStr)
		if err != nil {
			return false, 0, err
		}
	
		minutesLeft := math.Ceil(time.Until(blockedUntil).Minutes())
		if minutesLeft < 0 {
			minutesLeft = 0
		}
	
		return true, int64(minutesLeft), nil
	}
	
	return false, 0, nil
}

func (s *Students) LogLoginAttempt(email string, result bool) {
	attemptTime := time.Now().UTC()

	err := s.storage.LogAttempt(email, result, attemptTime)
	if err != nil {
		fmt.Printf("Ошибка логирования: %v\n", err)
	}
}

func (s *Students) GetFailedAttempts(email string) (int64, error) {
	now := time.Now().UTC()
	windowStart := now.Add(-1 * time.Minute)
	
	count, err := s.storage.GetFailedLogAttempts(email, windowStart)
	if err != nil {
		fmt.Printf("Ошибка подсчета попыток: %v\n", err)
		return int64(0), err
	}

	return int64(count), err
}

func (s *Students) BlockUser(email string) {
	now := time.Now()
	blockedUntil := now.Add(1 * time.Minute).Format(time.RFC3339)

	s.LogLoginAttempt(email, false)

	err := s.storage.BlockStudent(email, blockedUntil)
	if err != nil {
		fmt.Printf("Ошибка блокировки: %v\n", err)
	}
}