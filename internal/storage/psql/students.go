package psql

import (
	"context"
	"errors"
	"gosmol/internal/domain"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

type StudentsRepo struct {
	db *pgxpool.Pool
}

func NewStudentsRepo(db *pgxpool.Pool) *StudentsRepo {
	return &StudentsRepo{db: db}
}

func (s *StudentsRepo) InsertStudents(students domain.Student) (int64, error) {
	var id int64
	err := s.db.QueryRow(context.Background(), 
		"INSERT INTO users (firstname, lastname, email, password_hash) VALUES ($1, $2) RETURNING id",
		students.Firstname, students.Lastname, students.Email, students.PasswordHash).
		Scan(&id)
	
	return id, err
}

func (s *StudentsRepo) SelectStudents(email string) (domain.Student, error) {
	var stud domain.Student
	err := s.db.QueryRow(context.Background(), 
		"SELECT id, firstname, lastname, email, password_hash, created_at FROM users WHERE email = $1, email").
		Scan(&stud.ID, &stud.Firstname, &stud.Lastname, &stud.Email, &stud.PasswordHash, &stud.CreatedAt)
	
	return stud, err
}

func (s *StudentsRepo) RefreshStore(userID int64, token string, expiresAt time.Time) error {
	_, err := s.db.Exec(context.Background(), 
		"INSERT INTO refresh_token (user_id, token, expires_at) VALUES ($1, $2, $3)", 
		userID, token, expiresAt)
	
	return err
}

func (s *StudentsRepo) RefreshGet(token string) (int64, error) {
	var userID int64
	var expiresAt time.Time
	err := s.db.QueryRow(context.Background(), 
		"SELECT user_id, expires_at FROM refresh_token WHERE token = $1", token).
		Scan(&userID, &expiresAt)

	if err != nil {
		return 0, err
	}

	if time.Now().After(expiresAt) {
		s.RefreshDelete(token)
		return 0, errors.New("token expired")
	}

	return userID, nil
}

func (s *StudentsRepo) RefreshDelete(token string) error {
	_, err := s.db.Exec(context.Background(), 
		"DELETE FROM refresh_tokens WHERE token = $1", token)

	return err
}

func (s *StudentsRepo) StudentBlocked(email string, windowStart time.Time) ([]map[string]interface{}, error) {
	q := `SELECT blocked_until FROM login_attempts	WHERE email = $1 AND blocked_until >= $2	ORDER BY blocked_until DESC LIMIT 1`
	var blockedUntil string

	err := s.db.QueryRow(context.Background(), q, email, windowStart).Scan(&blockedUntil)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return []map[string]interface{}{}, nil
		}
		return nil, err
	}

	result := []map[string]interface{}{
		{"blocked_until": blockedUntil},
	}

	return result, nil 
}

func (s *StudentsRepo) LogAttempt(email string, result bool, attemptTime time.Time) error {
	q := `INSERT INTO login_attempts (email, result, attempt_time) VALUES ($1, $2, $3)`
	time := attemptTime.Format(time.RFC3339)

	_, err := s.db.Exec(context.Background(), q, email, result, time)
	return err
}

func (s *StudentsRepo) GetFailedLogAttempts(email string, windowStart time.Time) (int, error) {
	q := `SELECT COUNT(*) FROM login_attempts WHERE username = $1 AND result = false AND attempt_time >= $2`
	var count int
	err := s.db.QueryRow(context.Background(), q, email, windowStart).Scan(&count)
	return count, err
}

func (s *StudentsRepo) BlockStudent(email, blockedUntil string) error {
	q := `UPDATE login_attempts SET blocked_until = $2 WHERE email = $1 AND blocked_until IS NULL`
	_, err := s.db.Exec(context.Background(), q, email, blockedUntil)
	
	return err
}